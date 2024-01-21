package api

import (
	"afho__backend/botClient"

	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
	discordClient *botClient.BotClient
	server        *gin.Engine
}

func (handler *ApiHandler) InitAPI(discordClient *botClient.BotClient) {
	handler.discordClient = discordClient
	handler.server = gin.Default()
	handler.server.GET("/fetch", Fetch(discordClient))
	handler.server.GET("/connectedMembers", ConnectedMembers(discordClient))
	handler.server.GET("/brasilBoard", GetBrasilBoard(discordClient))
	handler.server.GET("/levels", GetLevels(discordClient))
	handler.server.GET("/times", GetTimes(discordClient))
	handler.server.Run("localhost:4000")
}
