package api

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	supa "github.com/nedpals/supabase-go"
	"github.com/realTristan/disgoauth"
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

func (handler *Handler) corsMiddleWare(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}

func (handler *Handler) checkUserMiddleware(c *gin.Context) {
	jwt := c.Request.Header.Get("Authorization")
	if jwt == "" {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Missing Authorization header",
		})
		return
	}

	accessToken, _, err := handler.GetTokenFromDB(jwt)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := disgoauth.GetUserData(accessToken)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Set("user", user)
	c.Next()
}

func (handler *Handler) checkAdminMiddleware(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	admins := GetAdmins(handler.discordClient.CacheHandler.Members, handler.discordClient.Config.AdminRoleID)

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

func (handler *Handler) GetTokenFromDB(jwt_token string) (string, string, error) {
	_, err := jwt.Parse(jwt_token, func(token *jwt.Token) (interface{}, error) {
		return []byte(handler.discordClient.Config.JWTSecret), nil
	})
	if err != nil {
		return "", "", err
	}

	var AccessToken string
	var RefreshToken string
	err = handler.discordClient.DB.QueryRow("SELECT access_token, refresh_token FROM discord_tokens WHERE jwt_token = ?", jwt_token).Scan(&AccessToken, &RefreshToken)
	if err != nil {
		return "", "", err
	}

	return AccessToken, RefreshToken, nil
}
