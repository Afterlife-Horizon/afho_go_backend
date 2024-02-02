package main

import (
	"afho__backend/api"
	"afho__backend/botClient"
	"afho__backend/utils"
	"context"
	"database/sql"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
var apiHandler *api.Handler = &api.Handler{}
var discordClient *botClient.BotClient = &botClient.BotClient{}

var tickerMinute *time.Ticker = time.NewTicker(time.Minute)
var env = utils.LoadEnv(utils.Flags{
	AddCommands: addCommandsFlag,
	DelCommands: delCommandsFlag,
})

var addCommandsFlag = flag.Bool("add-commands", false, "Add new commands to discord servers")
var delCommandsFlag = flag.Bool("del-commands", false, "Delete commands from discord servers")

func main() {
	flag.Parse()

	initDBConnection()
	initDiscordClient()
	go initAPI()

	go everyMinuteLoop()

	gracefulShutdown()
}

func everyMinuteLoop() {
	// <-discordClient.ReadyChannel
	for range tickerMinute.C {
		discordClient.CacheHandler.UpdateCache(discordClient)
	}
}

func initDiscordClient() {
	discordClient.Init(env, db)
}

func initAPI() {
	apiHandler.Init(discordClient)
}

func initDBConnection() {
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
		log.Fatalln(err.Error())
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	log.Println("DB Connected!")
}

func gracefulShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Gracefully Shutting Down!")

	tickerMinute.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiHandler.Server.Shutdown(ctx); err != nil {
		log.Fatalln(err.Error())
	}

	if err := discordClient.Session.Close(); err != nil {
		log.Fatalln(err.Error())
	}
}
