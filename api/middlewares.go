package api

import (
	"context"

	"github.com/gin-gonic/gin"
	supa "github.com/nedpals/supabase-go"
)

func (handler *Handler) readyMiddleware(c *gin.Context) {
	if !handler.discordClient.Ready {
		c.AbortWithStatusJSON(503, gin.H{
			"error": "Bot is not ready yet",
		})
		return
	}
	c.Next()
}

func (handler *Handler) checkUserMiddleware(c *gin.Context) {
	ok := true
	userToken := c.Request.Header.Get("Authorization")
	if userToken == "" {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Missing Authorization header",
		})
		ok = false
		return
	}

	user, err := handler.supabaseClient.Auth.User(context.Background(), userToken)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		ok = false
		return
	}

	if ok {
		c.Set("user", user)
		c.Next()
	}
}

func (handler *Handler) checkAdminMiddleware(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	admins := GetAdmins(handler.discordClient)

	var isAdmin bool
	for _, admin := range admins {
		if admin != user.UserMetadata["full_name"].(string) {
			continue
		}
		isAdmin = true
		c.Set("user", user)
		c.Next()
	}

	if !isAdmin {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "User is not admin",
		})
	}
}
