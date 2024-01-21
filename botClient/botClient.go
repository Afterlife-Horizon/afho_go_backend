package botClient

import (
	"afho__backend/utils"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type BotClient struct {
	Config          utils.Env
	DB              *sql.DB
	Session         *discordgo.Session
	CacheHandler    *CacheHandler
	MusicHandler    *MusicHandler
	VoiceHandler    *VoiceHandler
	CommandsBuilder *CommandsBuilder
	Commands        []*discordgo.ApplicationCommand
}

func (b *BotClient) Init(env utils.Env, db *sql.DB) {
	b.Config = env
	b.DB = db
	discord, err := discordgo.New("Bot " + env.Discord_token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	b.Session = discord
	discord.ShouldReconnectOnError = false
	discord.StateEnabled = true
	discord.LogLevel = discordgo.LogError

	discord.AddHandler(MessageCreate(b))
	discord.AddHandler(VoiceStateUpdate(b))
	discord.AddHandler(InteractionCreate(b))
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Ready to operate!")

		var cacheHandler CacheHandler
		cacheHandler.Init(b)
		log.Println("Initialised Cache Handler")

		b.CacheHandler = &cacheHandler

		var musicHandler MusicHandler
		musicHandler.Init(b)
		log.Println("Initialised Music Handler")

		b.MusicHandler = &musicHandler

		var voiceHandler = NewVoiceHandler(b)
		b.VoiceHandler = &voiceHandler

		var commandsBuilder = CommandsBuilder{}
		commandsBuilder.Init(b)

		b.CommandsBuilder = &commandsBuilder

		if *b.Config.DelCommands {
			b.CommandsBuilder.DeleteCommands(b)
		}
		if *b.Config.Flags.AddCommands {
			commandsBuilder.RegisterCommands(b)
		}

		if *b.Config.Flags.AddCommands || *b.Config.DelCommands {
			b.Session.Close()
			os.Exit(0)
		}
	})

	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("Started session as %v\n", discord.State.User.Username)
}

func (b *BotClient) Close() {
	b.Session.Close()
}

func (b *BotClient) BrazilUser(sender *discordgo.User, user *discordgo.User) string {
	// check if user is in voice channel and if senderis in the same voice channel
	var senderVoiceState, err = b.Session.State.VoiceState(b.Config.GuildID, sender.ID)
	if err != nil {
		log.Println(err.Error())
		return "Could not find sender voice state"
	}
	var userVoiceState, err2 = b.Session.State.VoiceState(b.Config.GuildID, user.ID)
	if err2 != nil {
		log.Println(err2.Error())
		return "Could not find user voice state"
	}
	if senderVoiceState.ChannelID != userVoiceState.ChannelID {
		b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v is not in the same voice channel as %v", user.Username, sender.Username))
		return "User is not in the same voice channel as sender"
	}

	// TODO: modify database record

	// check if user is already in brasil
	var isInBrasil = userVoiceState.ChannelID == b.Config.BrasilChannelID
	if isInBrasil {
		b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v is already in brasil", user.Username))
		return "User is already in brasil"
	}

	// move user to brasil
	b.Session.GuildMemberMove(b.Config.GuildID, user.ID, &b.Config.BrasilChannelID)
	b.Session.ChannelMessageSend(b.Config.BaseChannelID, fmt.Sprintf("%v sent %v to brasil!", sender.Username, user.Username))
	return "User sent to brasil"
}
