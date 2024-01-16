package main

import (
	"afho__backend/botClient"
	"afho__backend/utils"
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"
)

var addCommandsFlag = flag.Bool("add-commands", false, "Add new commands to discord servers")
var delCommandsFlag = flag.Bool("del-commands", false, "Delete commands from discord servers")

func main() {
	flag.Parse()
	var env = utils.LoadEnv(utils.Flags{
		AddCommands: addCommandsFlag,
		DelCommands: delCommandsFlag,
	})
	var Client botClient.BotClient
	Client.Init(env)

	ticker := time.NewTicker(time.Minute)
	go everyMinuteLoop(ticker, &Client)

	gracefulShutdown(&Client, *ticker)
}

func everyMinuteLoop(ticker *time.Ticker, client *botClient.BotClient) {
	for range ticker.C {
		client.CacheHandler.UpdateCache(client)
	}
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
