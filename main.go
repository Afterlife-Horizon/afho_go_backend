package main

import (
	"afho__backend/api"
	"afho__backend/botClient"
	"afho__backend/utils"
	"context"
	"database/sql"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
var apiHandler *api.Handler = &api.Handler{}
var discordClient *botClient.BotClient = &botClient.BotClient{}

var tickerMinute *time.Ticker = time.NewTicker(time.Minute)
var env utils.Env

var addCommandsFlag = flag.Bool("add-commands", false, "Add new commands to discord servers")
var delCommandsFlag = flag.Bool("del-commands", false, "Delete commands from discord servers")

func main() {
	flag.Parse()
	utils.InitLogger()
	env = utils.LoadEnv(utils.Flags{
		AddCommands: addCommandsFlag,
		DelCommands: delCommandsFlag,
	})

	initDBConnection()
	initDiscordClient()

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	utils.Logger.Debug("Shutting down API and Discord Client")
	if err := apiHandler.Server.Shutdown(ctx); err != nil {
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
