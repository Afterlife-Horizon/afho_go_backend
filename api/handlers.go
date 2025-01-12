package api

import (
	"afho_backend/botClient"
	"afho_backend/utils"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	supa "github.com/nedpals/supabase-go"
	"github.com/realTristan/disgoauth"
)

func (handler *Handler) getLevels(c *gin.Context) {
	var levels []Level = GetLevelsDb(handler.discordClient.DB, handler.discordClient.CacheHandler.Members)
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].Xp > levels[j].Xp
	})
	c.JSON(200, levels)
}

func (handler *Handler) generalFetch(c *gin.Context) {
	var prog time.Duration
	var tracks []Track
	var paused bool
	handler.discordClient.MusicHandler.Queue.RLock()
	if handler.discordClient.MusicHandler.Queue != nil && handler.discordClient.MusicHandler.Queue.Playing {
		prog = handler.discordClient.MusicHandler.Queue.SeekPosition + handler.discordClient.MusicHandler.Stream.PlaybackPosition()
		tracks = utils.Map(&handler.discordClient.MusicHandler.Queue.Tracks, func(track botClient.Track) Track {
			return Track{
				Id:                track.ID,
				Title:             track.Title,
				Duration:          int(track.Duration.Seconds()),
				DurationFormatted: utils.FormatTime(track.Duration),
				Requester:         track.RequestedBy,
				Thumbnail:         track.Thumbnail,
			}
		}).Data
		paused = handler.discordClient.MusicHandler.Paused
	} else {
		prog = 0
		tracks = []Track{}
		paused = false
	}

	handler.discordClient.MusicHandler.Queue.RUnlock()

	c.JSON(200, FetchResults{
		Admins:       GetAdmins(handler.discordClient.CacheHandler.Members, handler.discordClient.Config.AdminRoleID),
		Formatedprog: utils.FormatTime(prog),
		Prog:         int(prog.Seconds()),
		Queue: []Queue{
			{
				Effects: Effects{},
				Paused:  paused,
				Tracks:  tracks,
			},
		},
	})
}

func (handler *Handler) connectedMembers(c *gin.Context) {
	handler.discordClient.CacheHandler.VoiceConnectedMembers.RLock()
	var connectedMembers []connectedMembers = utils.Map[*discordgo.Member](handler.discordClient.CacheHandler.VoiceConnectedMembers, func(member *discordgo.Member) connectedMembers {
		return connectedMembers{
			Username: member.User.Username,
			ID:       member.User.ID,
		}
	}).Data
	handler.discordClient.CacheHandler.VoiceConnectedMembers.RUnlock()

	c.JSON(200, ConnectedMembersResponse{
		Data: connectedMembers,
	})
}

func (handler *Handler) getBrasilBoard(c *gin.Context) {
	var brasilBoard []BrasilBoard = GetBrasilBoardDB(handler.discordClient.DB, handler.discordClient.CacheHandler.Members)

	sort.Slice(brasilBoard, func(i, j int) bool {
		return brasilBoard[i].BresilReceived > brasilBoard[j].BresilReceived
	})

	c.JSON(200, brasilBoard)
}

func (handler *Handler) getTimes(c *gin.Context) {
	var times []Time = GetTimesDB(handler.discordClient.CacheHandler.Members, handler.discordClient.DB)

	sort.Slice(times, func(i, j int) bool {
		return times[i].TimeSpent > times[j].TimeSpent
	})

	c.JSON(200, times)
}

func (handler *Handler) getAchievements(c *gin.Context) {
	var achievements []APIAchievement = GetAchievementsDB(handler.discordClient.CacheHandler.Members, handler.discordClient.DB)
	c.JSON(200, achievements)
}

func (handler *Handler) getFavs(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	favs, err := GetFavsDB(handler.discordClient.DB, user.UserMetadata["provider_id"].(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"favorites": favs,
	})
}

func (handler *Handler) postPlay(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	body := struct {
		Songs string `json:"songs"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var songs []string = strings.Split(body.Songs, ",")

	for _, song := range songs {
		_, err := handler.discordClient.MusicHandler.Add(handler.discordClient, song, user.UserMetadata["full_name"].(string), user.UserMetadata["provider_id"].(string), false)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{})
}

func (handler *Handler) postClearQueue(c *gin.Context) {
	handler.discordClient.MusicHandler.ClearQueue(handler.discordClient)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postStop(c *gin.Context) {
	handler.discordClient.MusicHandler.Stop(handler.discordClient)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postSkip(c *gin.Context) {
	handler.discordClient.MusicHandler.Skip(handler.discordClient)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postDisconnect(c *gin.Context) {
	handler.discordClient.MusicHandler.Stop(handler.discordClient)
	botClient.HandleLeave(handler.discordClient)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postPause(c *gin.Context) {
	handler.discordClient.MusicHandler.SetPause(true)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postUnpause(c *gin.Context) {
	handler.discordClient.MusicHandler.SetPause(false)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postPlayFirst(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	body := struct {
		Songs string `json:"songs"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var songs []string = strings.Split(body.Songs, ",")

	for _, song := range songs {
		_, err := handler.discordClient.MusicHandler.Add(handler.discordClient, song, user.UserMetadata["full_name"].(string), user.UserMetadata["provider_id"].(string), true)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{})
}

func (handler *Handler) postSuffle(c *gin.Context) {
	handler.discordClient.MusicHandler.Shuffle(handler.discordClient)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postRemove(c *gin.Context) {
	body := struct {
		QueuePos int `json:"queuePos"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	handler.discordClient.MusicHandler.Remove(handler.discordClient, body.QueuePos)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postSkipTo(c *gin.Context) {
	body := struct {
		QueuePos int `json:"queuePos"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	handler.discordClient.MusicHandler.SkipTo(handler.discordClient, body.QueuePos)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postFilters(c *gin.Context) {
	body := struct {
		Filters Effects `json:"filters"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var filters string = body.Filters.ToFilters()

	handler.discordClient.MusicHandler.SetFilters(handler.discordClient, filters)
	c.JSON(200, gin.H{})
}

func (handler *Handler) postBresil(c *gin.Context) {
	_ = c.MustGet("user").(*supa.User)
	body := struct {
		MoverId string `json:"moverId"`
		MovedId string `json:"movedId"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	mover, err2 := handler.discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
		return t.User.ID == body.MoverId
	})

	if err2 != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err2.Error(),
		})
		return
	}

	moved, err3 := handler.discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
		return t.User.ID == body.MovedId
	})

	if err3 != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err3.Error(),
		})
		return
	}

	_, err := handler.discordClient.BrazilUser(mover.User, moved.User)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

func (handler *Handler) postAddFav(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	body := struct {
		Url string `json:"url"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err := AddFavDB(handler.discordClient.MusicHandler.YoutubeClient, handler.discordClient.DB, user.UserMetadata["provider_id"].(string), body.Url)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

func (handler *Handler) deleteRemoveFav(c *gin.Context) {
	user := c.MustGet("user").(*supa.User)
	body := struct {
		Id string `json:"id"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := RemoveFavoriteDB(handler.discordClient.DB, user.UserMetadata["provider_id"].(string), body.Id)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

func (handler *Handler) LoginHandler(c *gin.Context) {
	var prompt = ""
	if c.Query("prompt") != "" {
		prompt = "&prompt=" + c.Query("prompt")
	}

	var redirectURI = c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = handler.discordAuthClient.RedirectURI
	}

	var url = fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code%s",
		handler.discordAuthClient.ClientID,
		redirectURI,
		prompt,
	)

	// add scopes
	url += "&scope="
	for i, scope := range handler.discordAuthClient.Scopes {
		if i != len(handler.discordAuthClient.Scopes)-1 {
			url += scope + "%20"
			continue
		}
		url += scope
	}

	c.Redirect(302, url)
}

func (handler *Handler) CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Missing code",
		})
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = handler.discordAuthClient.RedirectURI
	}

	if redirectURI != handler.discordAuthClient.RedirectURI {
		handler.discordAuthClient.RedirectURI = redirectURI
	}

	accessToken, refresh_token, _, err := handler.discordAuthClient.GetAccessToken(code)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	// if expires_in < 10000 {
	// 	// Refresh token
	// 	data, err := handler.discordAuthClient.RefreshAccessToken(refresh_token)
	// 	if err != nil {
	// 		utils.Logger.Error(err.Error())
	// 		return
	// 	}

	// 	accessToken = data["access_token"].(string)
	// 	refresh_token = data["refresh_token"].(string)
	// }

	tmpUser, err := disgoauth.GetUserData(accessToken)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Could not get user data",
		})
		return
	}

	user := DiscordUserFromMap(tmpUser)

	randomString := utils.RandomString(32)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Issuer":       "Afterlife Horizon Bot",
		"user":         user.Id,
		"randomString": randomString,
	})
	signedToken, err := token.SignedString([]byte(handler.discordClient.Config.JWTSecret))
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	c.JSON(200, gin.H{
		"token": signedToken,
		"user":  user,
	})

	handler.SaveTokenIntoDB(user, accessToken, refresh_token, signedToken)
}

func (handler *Handler) SaveTokenIntoDB(user *DiscordApiUser, accessToken, refreshToken string, jwtToken string) {
	_, err := handler.discordClient.DB.Exec("INSERT INTO discord_tokens (user_id, access_token, refresh_token, jwt_token) VALUES (?, ?, ?, ?)", user.Id, accessToken, refreshToken, jwtToken)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
}

func (handler *Handler) GetUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}
