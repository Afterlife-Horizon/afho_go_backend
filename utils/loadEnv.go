package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	Discord_token string
	GuildID       string
	YTApiKey      string
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
		log.Fatalln("YTApiKey not found in environment variables")
	}

	return Env{
		Discord_token: discord_token,
		GuildID:       guildID,
		YTApiKey:      YTApiKey,
		Flags:         flags,
	}
}
