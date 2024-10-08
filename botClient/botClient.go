package botClient

import (
	"afho_backend/utils"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

type BotClient struct {
	ReadyChannel    chan bool
	DB              *sql.DB
	Session         *discordgo.Session
	CacheHandler    *CacheHandler
	MusicHandler    *MusicHandler
	VoiceHandler    *VoiceHandler
	CommandsBuilder *CommandsBuilder
	Config          utils.Env
	Ready           bool
}

func (b *BotClient) Init(env utils.Env, db *sql.DB) {
	utils.Logger.Debug("Initialising Discord Client")
	b.Config = env
	b.DB = db

	utils.Logger.Debug("Creating Discord Session")
	discord, err := discordgo.New("Bot " + env.Discord_token)
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}
	b.Session = discord

	discord.ShouldReconnectOnError = false
	discord.StateEnabled = true
	discord.LogLevel = discordgo.LogError
	utils.Logger.Debug("Discord bot variables set to:", "StateEnabled", discord.StateEnabled, "LogLevel", discord.LogLevel, "ShouldReconnectOnError", discord.ShouldReconnectOnError)

	utils.Logger.Debug("Adding Discord Event Handlers")
	discord.AddHandler(MessageCreate(b))
	discord.AddHandler(VoiceStateUpdate(b))
	discord.AddHandler(InteractionCreate(b))
	discord.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) {
		utils.Logger.Debug("Discord Bot Ready Event Received")
		b.ReadyChannel = make(chan bool)

		var cacheHandler CacheHandler
		cacheHandler.Init(b)
		utils.Logger.Info("Initialised Cache Handler")

		b.CacheHandler = &cacheHandler

		var musicHandler MusicHandler
		musicHandler.Init(b)
		utils.Logger.Info("Initialised Music Handler")

		b.MusicHandler = &musicHandler

		var voiceHandler VoiceHandler
		voiceHandler.Init(b)
		b.VoiceHandler = &voiceHandler
		utils.Logger.Info("Initialised Voice Handler")

		go musicHandler.HandleQueue(b)

		commandsBuilder := CommandsBuilder{}
		commandsBuilder.Init(b)
		utils.Logger.Info("Initialised Commands Builder")

		b.CommandsBuilder = &commandsBuilder

		if *b.Config.DelCommands {
			b.CommandsBuilder.DeleteCommands(b)
		}

		if *b.Config.AddCommands {
			commandsBuilder.RegisterCommands(b)
		}

		if *b.Config.AddCommands || *b.Config.DelCommands {
			utils.Logger.Debug("Exiting after adding or deleting commands")
			b.Session.Close()
			os.Exit(0)
		}

		b.Ready = true
		b.ReadyChannel <- true
		utils.Logger.Info("Discord Bot Ready!")

		b.CacheHandler.UpdateCache()
	})

	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	utils.Logger.Infof("Started session as %v\n", discord.State.User.Username)
}

func (b *BotClient) Close() {
	b.Session.Close()
}

const (
	bresil_sent = iota
	bresil_recieved
)

func (b *BotClient) BrazilUser(sender *discordgo.User, user *discordgo.User) (string, error) {
	// check if user is in voice channel and if senders in the same voice channel
	senderVoiceState, err := b.Session.State.VoiceState(b.Config.GuildID, sender.ID)
	if err != nil {
		utils.Logger.Error(err.Error())
		return "Could not find sender voice state", err
	}
	userVoiceState, err2 := b.Session.State.VoiceState(b.Config.GuildID, user.ID)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return "Could not find user voice state", err2
	}
	if senderVoiceState.ChannelID != userVoiceState.ChannelID {
		b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v is not in the same voice channel as %v", user.Username, sender.Username))
		return "User is not in the same voice channel as sender", errors.New("user is not in the same voice channel as sender")
	}

	// check if user is already in brasil
	isInBrasil := userVoiceState.ChannelID == b.Config.BrasilChannelID
	if isInBrasil {
		b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v is already in brasil", user.Username))
		return "User is already in brasil", errors.New("user is already in brasil")
	}

	// move user to brasil
	b.Session.GuildMemberMove(b.Config.GuildID, user.ID, &b.Config.BrasilChannelID)
	b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v sent %v to brasil!", sender.Username, user.Username))

	bresilUser(b.DB, user, bresil_recieved)
	bresilUser(b.DB, sender, bresil_sent)

	return "User sent to brasil", nil
}

func bresilUser(DB *sql.DB, user *discordgo.User, updateType int) (string, error) {
	var stmt *sql.Stmt
	var err error
	if updateType == bresil_sent {
		stmt, err = DB.Prepare("INSERT INTO Bresil_count (user_id, bresil_received) VALUES (?, ?) ON DUPLICATE KEY UPDATE bresil_received = bresil_received+1")
		if err != nil {
			utils.Logger.Error(err.Error())
			return "Could not prepare statement", err
		}
	}
	if updateType == bresil_recieved {
		stmt, err = DB.Prepare("INSERT INTO Bresil_count (user_id, bresil_sent) VALUES (?, ?) ON DUPLICATE KEY UPDATE bresil_sent = bresil_sent+1")
		if err != nil {
			utils.Logger.Error(err.Error())
			return "Could not prepare statement", err
		}
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.ID, 1)
	if err != nil {
		utils.Logger.Error(err.Error())
		return "Could not execute statement", err
	}

	return "User sent to brasil", nil
}

func (b *BotClient) UpdateDB() {
	b.CacheHandler.UpdateDB()
	b.VoiceHandler.UpdateDBTime()
}
