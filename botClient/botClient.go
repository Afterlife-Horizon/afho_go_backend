package botClient

import (
	"afho__backend/utils"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type BotClient struct {
	Config       utils.Env
	Discord      *discordgo.Session
	CacheHandler *CacheHandler
}

func (b *BotClient) Init(env utils.Env) {
	b.Config = env
	discord, err := discordgo.New("Bot " + env.Discord_token)
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}
	discord.ShouldReconnectOnError = false

	discord.AddHandler(MessageCreate(b))
	discord.AddHandler(VoiceStateUpdate(b))

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
		b.Discord = discord

		var cacheHandler CacheHandler
		cacheHandler.Init(b)
		log.Println("Started Cache Handler")

		b.CacheHandler = &cacheHandler
	})

	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Started session as %v\n", discord.State.User.Username)
}
