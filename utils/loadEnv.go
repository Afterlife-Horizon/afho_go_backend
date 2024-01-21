package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	// Tokens
	Discord_token string
	YTApiKey      string

	// IDs
	GuildID         string
	BaseChannelID   string
	BrasilChannelID string

	// Flags
	Flags
}

type Flags struct {
	AddCommands *bool
	DelCommands *bool
}

func LoadEnv(flags Flags) Env {
	godotenv.Load()
	discord_token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		log.Fatalln("DISCORD_TOKEN not found in environment variables")
	}

	guildID, ok := os.LookupEnv("GUILD_ID")
	if !ok {
		log.Fatalln("GUILD_ID not found in environment variables")
	}

	YTApiKey, ok := os.LookupEnv("YT_API_KEY")
	if !ok {
		log.Fatalln("YT_API_KEY not found in environment variables")
	}

	BaseChannelID, ok := os.LookupEnv("BASE_CHANNEL_ID")
	if !ok {
		log.Fatalln("BASE_CHANNEL_ID not found in environment variables")
	}

	BrasilChannelID, ok := os.LookupEnv("BRASIL_CHANNEL_ID")
	if !ok {
		log.Fatalln("BRASIL_CHANNEL_ID not found in environment variables")
	}

	return Env{
		Discord_token:   discord_token,
		GuildID:         guildID,
		YTApiKey:        YTApiKey,
		BaseChannelID:   BaseChannelID,
		BrasilChannelID: BrasilChannelID,
		Flags:           flags,
	}
}
