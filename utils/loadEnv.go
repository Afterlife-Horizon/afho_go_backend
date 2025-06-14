package utils

import (
	"os"
)

type Env struct {
	Flags
	AdminRoleID     string
	DbUser          string
	Discord_token   string
	YTApiKey        string
	GuildID         string
	BaseChannelID   string
	KeyFile         string
	CertFile        string
	BrasilChannelID string
	DbPass          string
	DbName          string
	DbAddress       string

	ClientID     string
	ClientSecret string
	RedirectURI  string

	JWTSecret string

	IsProduction bool
}

type Flags struct {
	AddCommands *bool
	DelCommands *bool
}

func LoadEnv(flags Flags) Env {
	isProduction, ok := os.LookupEnv("IS_PRODUCTION")
	if !ok || (isProduction != "true" && isProduction != "false") {
		Logger.Warn("IS_PRODUCTION not found in environment variables, setting to false")
		isProduction = "false"
	}

	certFilePath, ok := os.LookupEnv("CERT_FILE")
	if !ok || certFilePath == "" {
		Logger.Warn("CERT_FILE not found in environment variables, not using HTTPS")
	}

	keyFilePath, ok := os.LookupEnv("KEY_FILE")
	if (!ok || keyFilePath == "") && certFilePath != "" {
		Logger.Fatal("KEY_FILE not found in environment variables")
	}

	discord_token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok || discord_token == "" {
		Logger.Fatal("DISCORD_TOKEN not found in environment variables")
	}

	guildID, ok := os.LookupEnv("GUILD_ID")
	if !ok || guildID == "" {
		Logger.Fatal("GUILD_ID not found in environment variables")
	}

	YTApiKey, ok := os.LookupEnv("YT_API_KEY")
	if !ok || YTApiKey == "" {
		Logger.Fatal("YT_API_KEY not found in environment variables")
	}

	BaseChannelID, ok := os.LookupEnv("BASE_CHANNEL_ID")
	if !ok || BaseChannelID == "" {
		Logger.Fatal("BASE_CHANNEL_ID not found in environment variables")
	}

	BrasilChannelID, ok := os.LookupEnv("BRASIL_CHANNEL_ID")
	if !ok || BrasilChannelID == "" {
		Logger.Fatal("BRASIL_CHANNEL_ID not found in environment variables")
	}

	DbAddress, ok := os.LookupEnv("DB_ADDRESS")
	if !ok || DbAddress == "" {
		Logger.Fatal("DB_ADDRESS not found in environment variables")
	}

	DbName, ok := os.LookupEnv("DB_NAME")
	if !ok || DbName == "" {
		Logger.Fatal("DB_NAME not found in environment variables")
	}

	DbUser, ok := os.LookupEnv("DB_USER")
	if !ok || DbUser == "" {
		Logger.Fatal("DB_USER not found in environment variables")
	}

	DbPass, ok := os.LookupEnv("DB_PASS")
	if !ok || DbPass == "" {
		Logger.Fatal("DB_PASS not found in environment variables")
	}

	ClientID, ok := os.LookupEnv("DISCORD_CLIENT_ID")
	if !ok || ClientID == "" {
		Logger.Fatal("DISCORD_CLIENT_ID not found in environment variables")
	}

	ClientSecret, ok := os.LookupEnv("DISCORD_CLIENT_SECRET")
	if !ok || ClientSecret == "" {
		Logger.Fatal("DISCORD_CLIENT_SECRET not found in environment variables")
	}

	RedirectURI, ok := os.LookupEnv("DISCORD_REDIRECT_URI")
	if !ok || RedirectURI == "" {
		Logger.Fatal("DISCORD_REDIRECT_URI not found in environment variables")
	}

	JWTSecret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || RedirectURI == "" {
		Logger.Fatal("JWT_SECRET not found in environment variables")
	}

	AdminRoleID, ok := os.LookupEnv("ADMIN_ROLE_ID")
	if !ok || AdminRoleID == "" {
		Logger.Fatal("ADMIN_ROLE_ID not found in environment variables")
	}

	return Env{
		IsProduction:    isProduction == "true",
		CertFile:        certFilePath,
		KeyFile:         keyFilePath,
		Discord_token:   discord_token,
		GuildID:         guildID,
		YTApiKey:        YTApiKey,
		BaseChannelID:   BaseChannelID,
		BrasilChannelID: BrasilChannelID,
		AdminRoleID:     AdminRoleID,
		DbAddress:       DbAddress,
		DbName:          DbName,
		DbUser:          DbUser,
		DbPass:          DbPass,
		ClientID:        ClientID,
		ClientSecret:    ClientSecret,
		RedirectURI:     RedirectURI,
		JWTSecret:       JWTSecret,
		Flags:           flags,
	}
}
