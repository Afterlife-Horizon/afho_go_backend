package botClient

import (
	"afho__backend/utils"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type BotClient struct {
	Config          utils.Env
	Discord         *discordgo.Session
	CacheHandler    *CacheHandler
	MusicHandler    *MusicHandler
	CommandsBuilder *CommandsBuilder
	Commands        []*discordgo.ApplicationCommand
}

func (b *BotClient) Init(env utils.Env) {
	b.Config = env
	discord, err := discordgo.New("Bot " + env.Discord_token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	b.Discord = discord
	discord.ShouldReconnectOnError = false
	discord.StateEnabled = true
	discord.LogLevel = discordgo.LogError

	discord.AddHandler(MessageCreate(b))
	discord.AddHandler(VoiceStateUpdate(b))
	discord.AddHandler(InteractionCreate(b))
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Ready to operate!")

		var commandsBuilder = CommandsBuilder{}
		commandsBuilder.Init(b)

		b.CommandsBuilder = &commandsBuilder

		if *b.Config.Flags.AddCommands {
			commandsBuilder.RegisterCommands(b)
		}

		// boken
		if *b.Config.DelCommands {
			b.CommandsBuilder.DeleteCommands(b)
		}

		var cacheHandler CacheHandler
		cacheHandler.Init(b)
		log.Println("Initialised Cache Handler")

		b.CacheHandler = &cacheHandler

		var musicHandler MusicHandler
		musicHandler.Init(b)
		log.Println("Initialised Music Handler")

		b.MusicHandler = &musicHandler
	})

	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("Started session as %v\n", discord.State.User.Username)
}
