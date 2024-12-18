package main

import (
	"afho_backend/api"
	"afho_backend/botClient"
	"afho_backend/utils"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	db            *sql.DB
	apiHandler    *api.Handler         = &api.Handler{}
	discordClient *botClient.BotClient = &botClient.BotClient{}
)

var (
	tickerMinute    *time.Ticker  = time.NewTicker(time.Minute)
	tickerTenSecond *time.Ticker  = time.NewTicker(10 * time.Second)
	uptime          time.Duration = 0
	env             utils.Env
)

var (
	addCommandsFlag = flag.Bool("add-commands", false, "Add new commands to discord servers")
	delCommandsFlag = flag.Bool("del-commands", false, "Delete commands from discord servers")
	debugFlag       = flag.Bool("debug", false, "Run in debug mode")
)

func main() {
	flag.Parse()
	_ = godotenv.Load()

	utils.InitLogger(*debugFlag)
	env = utils.LoadEnv(utils.Flags{
		AddCommands: addCommandsFlag,
		DelCommands: delCommandsFlag,
	})

	initDBConnection()
	initDiscordClient()

	go everyTenSecondLoop()
	go everyMinuteLoop()
	go initAPI()

	gracefulShutdown()
}

func everyMinuteLoop() {
	// <-discordClient.ReadyChannel
	for range tickerMinute.C {
		utils.Logger.Debug("Cache and DB update loop run!")
		discordClient.CacheHandler.UpdateCache()
		discordClient.CacheHandler.UpdateDB()
	}
}

func everyTenSecondLoop() {
	for range tickerTenSecond.C {
		uptime = uptime + time.Second*10
		utils.Logger.Debug("Uptime:", uptime)
		formattedUptime := utils.FormatTime(uptime)
		err := discordClient.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: fmt.Sprintf("Uptime: %v", formattedUptime),
					Type: discordgo.ActivityTypeWatching,
				},
			},
			Status: "online",
		})
		if err != nil {
			utils.Logger.Error(err.Error())
		}
	}
}

func initDiscordClient() {
	discordClient.Init(env, db)
}

func initAPI() {
	apiHandler.Init(discordClient)
}

func initDBConnection() {
	utils.Logger.Debug("Initialising DB connection")
	cfg := mysql.Config{
		User:                 env.DbUser,
		Passwd:               env.DbPass,
		Net:                  "tcp",
		Addr:                 env.DbAddress,
		DBName:               env.DbName,
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	pingErr := db.Ping()
	if pingErr != nil {
		utils.Logger.Fatal(pingErr)
	}
	utils.Logger.Info("DB Connected!")
}

func gracefulShutdown() {
	utils.Logger.Debug("Setting up graceful shutdown")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	utils.Logger.Info("Gracefully Shutting Down!")

	tickerMinute.Stop()
	tickerTenSecond.Stop()

	utils.Logger.Debug("Shutting down API and Discord Client")
	if err := apiHandler.Server.Shutdown(context.Background()); err != nil {
		utils.Logger.Fatal(err.Error())
	}

	utils.Logger.Debug("Shutting down Discord Client")
	utils.Logger.Debug("Updating DB values")
	discordClient.UpdateDB()

	if err := discordClient.Session.Close(); err != nil {
		utils.Logger.Fatal(err.Error())
	}

	utils.Logger.Debug("Closing DB connection")
	db.Close()
}
