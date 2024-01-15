package main

import (
	"afho__backend/botClient"
	"afho__backend/utils"
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var env = utils.LoadEnv()
	var botClient botClient.BotClient
	botClient.Init(env)

	ticker := time.NewTicker(time.Minute)
	go everyMinuteLoop(ticker, &botClient)

	gracefulShutdown(&botClient, *ticker)
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
	log.Println("got interruption signal")
	if err := client.Discord.Close(); err != nil {
		log.Fatalln(err.Error())
	}

	ticker.Stop()
}
