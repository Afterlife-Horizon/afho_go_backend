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

var addCommandsFlag = flag.Bool("add-commands", false, "Add new commands to discord servers")
var delCommandsFlag = flag.Bool("del-commands", false, "Delete commands from discord servers")

func main() {
	flag.Parse()
	var env = utils.LoadEnv(utils.Flags{
		AddCommands: addCommandsFlag,
		DelCommands: delCommandsFlag,
	})

	initDBConnection(env)

	Client := initDiscordClient(env)

	ticker := time.NewTicker(time.Minute)
	go everyMinuteLoop(ticker, Client)

	initAPI(Client)

	gracefulShutdown(Client, *ticker)
}

func everyMinuteLoop(ticker *time.Ticker, client *botClient.BotClient) {
	for range ticker.C {
		client.CacheHandler.UpdateCache(client)
	}
}

func initDiscordClient(env utils.Env) *botClient.BotClient {
	var Client botClient.BotClient
	Client.Init(env, db)
	return &Client
}

func initAPI(discodClient *botClient.BotClient) {
	var apiHandler api.ApiHandler
	apiHandler.InitAPI(discodClient)
}

func initDBConnection(env utils.Env) {
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
	log.Println("Connected!")
}

func gracefulShutdown(client *botClient.BotClient, ticker time.Ticker) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Gracefully Shutting Down!")

	if err := client.Session.Close(); err != nil {
		log.Fatalln(err.Error())
	}

	ticker.Stop()
}
