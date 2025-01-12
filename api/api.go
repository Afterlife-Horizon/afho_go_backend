package api

import (
	"afho_backend/botClient"
	"afho_backend/utils"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	disgoauth "github.com/realTristan/disgoauth"
)

type Handler struct {
	discordClient     *botClient.BotClient
	router            *gin.Engine
	discordAuthClient *disgoauth.Client
	Server            *http.Server
}

func (handler *Handler) Init(discordClient *botClient.BotClient) {
	utils.Logger.Debug("Initialising API handler")
	handler.discordClient = discordClient
	utils.Logger.Debug("Creating supabase client")

	handler.discordAuthClient = disgoauth.Init(&disgoauth.Client{
		ClientID:     discordClient.Config.ClientID,
		ClientSecret: discordClient.Config.ClientSecret,
		RedirectURI:  discordClient.Config.RedirectURI,
		Scopes:       []string{disgoauth.ScopeIdentify},
	})

	handler.setMode()

	utils.Logger.Debug("Waiting for bot to be ready")
	<-discordClient.ReadyChannel // Wait for bot to be ready

	handler.initRouter()

	handler.Server = &http.Server{
		Addr:    ":4000",
		Handler: handler.router,
	}

	utils.Logger.Debug("API Starting...")
	handler.run()
}

func (handler *Handler) initRouter() {
	utils.Logger.Debug("Initialising API router")
	handler.router = gin.Default()
	handler.router.Use(handler.readyMiddleware)
	handler.router.Use(handler.corsMiddleWare)

	handler.router.GET("/connectedMembers", handler.connectedMembers)
	handler.router.GET("/brasilBoard", handler.getBrasilBoard)
	handler.router.GET("/levels", handler.getLevels)
	handler.router.GET("/times", handler.getTimes)
	handler.router.GET("/achievements", handler.getAchievements)
	handler.router.GET("/getFavs", handler.checkUserMiddleware, handler.getFavs)
	handler.router.POST("/bresilMember", handler.checkUserMiddleware, handler.postBresil)

	// ------ Discord Auth Routes ------
	handler.router.GET("/discord/login", handler.LoginHandler)
	handler.router.GET("/discord/callback", handler.CallbackHandler)

	// ------ User Routes ------
	userRoutes := handler.router.Group("/user")
	userRoutes.Use(handler.checkUserMiddleware)
	userRoutes.GET("", handler.GetUser)

	// ------ Music Routes ------
	musicGroup := handler.router.Group("/music")
	musicGroup.GET("/fetch", handler.generalFetch)

	// --- USER ---

	musicGroup.POST("/play", handler.checkUserMiddleware, handler.postPlay)
	musicGroup.POST("/skip", handler.checkUserMiddleware, handler.postSkip)
	musicGroup.POST("/pause", handler.checkUserMiddleware, handler.postPause)
	musicGroup.POST("/unpause", handler.checkUserMiddleware, handler.postUnpause)
	musicGroup.POST("/playfirst", handler.checkUserMiddleware, handler.postPlayFirst)
	musicGroup.POST("/shuffle", handler.checkUserMiddleware, handler.postSuffle)
	musicGroup.POST("/addFav", handler.checkUserMiddleware, handler.postAddFav)
	musicGroup.DELETE("/delFav", handler.checkUserMiddleware, handler.deleteRemoveFav)

	// --- ADMIN ---
	musicGroup.POST("/remove", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postRemove)
	musicGroup.POST("/skipto", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postSkipTo)
	musicGroup.POST("/clearqueue", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postClearQueue)
	musicGroup.POST("/stop", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postStop)
	musicGroup.POST("/disconnect", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postDisconnect)
	musicGroup.POST("/filters", handler.checkUserMiddleware, handler.checkAdminMiddleware, handler.postFilters)
}

func (handler *Handler) setMode() {
	if gin.Mode() != gin.ReleaseMode && handler.discordClient.Config.IsProduction {
		utils.Logger.Debug("Setting gin mode to release")
		gin.SetMode(gin.ReleaseMode)
		return
	}

	if gin.Mode() != gin.DebugMode && !handler.discordClient.Config.IsProduction {
		utils.Logger.Debug("Setting gin mode to debug")
		gin.SetMode(gin.DebugMode)
	}
}

func (handler *Handler) run() {
	if handler.discordClient.Config.CertFile != "" && handler.discordClient.Config.KeyFile != "" {
		utils.Logger.Info("Starting HTTPS server")
		cwd, err := os.Getwd()
		if err != nil {
			utils.Logger.Fatal(err.Error())
		}
		certfile := handler.discordClient.Config.CertFile
		if certfile[0:1] != "/" {
			certfile = path.Join(cwd, handler.discordClient.Config.CertFile)
		}
		keyfile := handler.discordClient.Config.KeyFile
		if keyfile[0:1] != "/" {
			keyfile = path.Join(cwd, handler.discordClient.Config.KeyFile)
		}
		if err := handler.Server.ListenAndServeTLS(certfile, keyfile); err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatal(err.Error())
		}
		return
	}

	utils.Logger.Info("Starting HTTP server")
	if err := handler.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.Logger.Fatal(err.Error())
	}
}
