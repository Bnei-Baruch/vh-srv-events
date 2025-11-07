package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hellofresh/health-go/v5"
	healthpgx "github.com/hellofresh/health-go/v5/checks/pgx4"

	"gitlab.bbdev.team/vh/vh-srv-events/api/middleware"
	"gitlab.bbdev.team/vh/vh-srv-events/common"
	"gitlab.bbdev.team/vh/vh-srv-events/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

type App struct {
	repo      repo.EventsRepository
	eventsAPI *EventsAPI
	gEngine   *gin.Engine
}

func NewApp() *App {
	return new(App)
}

func (a *App) Initialize() {
	a.initSentry()
	a.initDB()
	a.eventsAPI = NewEventsAPI(a.repo)
	a.initGinEngine()
	a.initHealth()
}

func (a *App) initDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	a.repo, err = repo.NewEventsDB(ctx)
	if err != nil {
		utils.LogFatal("connect to db", slog.Any("err", err))
	}

	err = repo.SyncDBStructInsertionAndMigrations()
	if err != nil {
		utils.LogFatal("db migrations", slog.Any("err", err))
	}

	slog.Info("db connected and migrated")
}

func (a *App) initSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Release:          common.GitSHA,
		Environment:      common.Config.Env,
		AttachStacktrace: true,
	})
	if err != nil {
		utils.LogFatal("sentry.Init", slog.Any("err", err))
	}
}

func (a *App) initGinEngine() {
	gin.SetMode(common.Config.Mode)
	a.gEngine = gin.New()
	issuerUrl := fmt.Sprintf("%s/auth/realms/%s", common.Config.KeycloakServerUrl, common.Config.KeycloakRealm)
	tokenVerifier, err := middleware.NewFailoverOIDCTokenVerifier(issuerUrl)
	if err != nil {
		utils.LogFatal("middleware.NewFailoverOIDCTokenVerifier", slog.Any("err", err))
	}

	// middleware
	a.gEngine.Use(
		middleware.Logging(),
		middleware.Recovery(),
		sentrygin.New(sentrygin.Options{Repanic: true}),
		middleware.Sentry(),
	)
	if gin.IsDebugging() {
		a.gEngine.Use(cors.New(cors.Config{
			AllowAllOrigins:  true,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}
	a.gEngine.Use(
		middleware.TokenSource(),
		middleware.Authentication(tokenVerifier),
	)

	// routes
	basePath := a.gEngine.Group("/v1")

	participant := basePath.Group("/participant")
	{
		participant.POST("/", a.eventsAPI.CreateNewParticipant)
		participant.PATCH("/:id", a.eventsAPI.UpdateParticipantByID)
		participant.DELETE("/:id", a.eventsAPI.DeleteParticipantByID)
		participant.GET("/:id", a.eventsAPI.GetParticipantById)
		participant.GET("email/:email", a.eventsAPI.GetParticipantByEmail)
		participant.GET("keycloakid/:id", a.eventsAPI.GetParticipantByKeycloakID)
	}
	basePath.GET("/participants", a.eventsAPI.GetAllParticipant)

	participationOption := basePath.Group("/participation-option")
	{
		participationOption.POST("/", a.eventsAPI.CreateNewParticipationOption)
		participationOption.PATCH("/:name", a.eventsAPI.UpdateParticipationOptionByName)
		participationOption.DELETE("/:name", a.eventsAPI.DeleteParticipationOptionByName)
		participationOption.GET("/:name", a.eventsAPI.GetParticipationOptionByName)
	}
	basePath.GET("/participation-options", a.eventsAPI.GetAllParticipationOption)

	participationStatus := basePath.Group("/participation-status")
	{
		participationStatus.POST("", a.eventsAPI.CreateNewParticipationStatus)
		participationStatus.GET("/:id", a.eventsAPI.GetParticipationStatusByID)
		participationStatus.PATCH("/kcid/:kcid/event_slug/:slug", a.eventsAPI.UpdateParticipationStatus)
		participationStatus.PATCH("/:id", a.eventsAPI.UpdateParticipationStatus)
		participationStatus.DELETE("/:id", a.eventsAPI.DeleteParticipationStatusByID)
	}
	basePath.GET("/participation-statuses", a.eventsAPI.GetAllParticipationStatus)

	platform := basePath.Group("/platform")
	{
		platform.POST("/", a.eventsAPI.CreateNewPlatform)
		platform.PATCH("/:name", a.eventsAPI.UpdatePlatformByName)
		platform.DELETE("/:name", a.eventsAPI.DeletePlatformByName)
		platform.GET("/:name", a.eventsAPI.GetPlatformByName)
	}
	basePath.GET("/platforms", a.eventsAPI.GetAllPlatform)

	audience := basePath.Group("/audience")
	{
		audience.POST("/", a.eventsAPI.CreateNewAudience)
		audience.PATCH("/:name", a.eventsAPI.UpdateAudienceByName)
		audience.DELETE("/:name", a.eventsAPI.DeleteAudienceByName)
		audience.GET("/:name", a.eventsAPI.GetAudienceByName)
	}
	basePath.GET("/audiences", a.eventsAPI.GetAllAudience)

	broadcastURL := basePath.Group("/broadcasturl")
	{
		broadcastURL.POST("/", a.eventsAPI.CreateNewBroadcastURL)
		broadcastURL.PATCH("/:id", a.eventsAPI.UpdateBroadcastURLByID)
		broadcastURL.DELETE("/:id", a.eventsAPI.DeleteBroadcastURLByID)
		broadcastURL.GET("/:id", a.eventsAPI.GetBroadcastURLByID)
	}
	basePath.GET("/broadcasturls", a.eventsAPI.GetAllBroadcastURL)

	item := basePath.Group("/item")
	{
		item.POST("/", a.eventsAPI.CreateNewItem)
		item.GET("/:id", a.eventsAPI.GetItemByID)
		item.PATCH("/:id", a.eventsAPI.UpdateItemByID)
		item.DELETE("/:id", a.eventsAPI.DeleteItemByID)
	}
	basePath.GET("/items", a.eventsAPI.GetAllItem)

	itemBroadcastUrl := basePath.Group("/item-broadcasturl")
	{
		itemBroadcastUrl.POST("/", a.eventsAPI.CreateNewItemBroadcastURL)
		itemBroadcastUrl.GET("/:id", a.eventsAPI.GetItemBroadcastURLByID)
		itemBroadcastUrl.PATCH("/:id", a.eventsAPI.UpdateItemBroadcastURLByID)
		itemBroadcastUrl.DELETE("/:id", a.eventsAPI.DeleteItemBroadcastURLByID)
	}
	basePath.GET("/item-broadcasturls", a.eventsAPI.GetAllItemBroadcastURL)

	event := basePath.Group("/event")
	{
		event.POST("/", a.eventsAPI.CreateNewEvent)
		event.GET("/:id", a.eventsAPI.GetEventByID)
		event.PATCH("/:id", a.eventsAPI.UpdateEventByID)
		event.DELETE("/:id", a.eventsAPI.DeleteEventByID)
		event.DELETE("/hard/:id", a.eventsAPI.DeleteHardEventByID)
	}
	basePath.GET("/events", a.eventsAPI.GetAllEvent)

	eventItem := basePath.Group("/event-item")
	{
		eventItem.POST("/", a.eventsAPI.CreateNewEventItem)
		eventItem.GET("/:id", a.eventsAPI.GetEventItemByID)
		eventItem.PATCH("/:id", a.eventsAPI.UpdateEventItemByID)
		eventItem.DELETE("/:id", a.eventsAPI.DeleteEventItemByID)
	}
	basePath.GET("/event-items", a.eventsAPI.GetAllEventItem)

	eventPartOption := basePath.Group("/event-part-option")
	{
		eventPartOption.POST("/", a.eventsAPI.CreateNewEventPartOption)
		eventPartOption.GET("/:id", a.eventsAPI.GetEventPartOptionByID)
		eventPartOption.PATCH("/:id", a.eventsAPI.UpdateEventPartOptionByID)
		eventPartOption.DELETE("/:id", a.eventsAPI.DeleteEventPartOptionByID)
	}
	basePath.GET("/event-part-options", a.eventsAPI.GetAllEventPartOption)

	// operation := basePath.Group("/operation")
	// {
	// 	operation.POST("/", r.operationTrace.HandleOperationCreate)
	// 	operation.POST("/revert", r.operationTrace.HandleOperationRevert)
	// }

	emailNotification := basePath.Group("/notification")
	{
		emailNotification.POST("/event", a.eventsAPI.SendEventEmail)
	}

	analytics := basePath.Group("/analytics")
	{
		analytics.GET("/participants", a.eventsAPI.PartAnalytics)
	}
}

func (a *App) initHealth() {
	h, _ := health.New(health.WithComponent(health.Component{
		Name:    common.ServiceName,
		Version: common.GitSHA,
	}), health.WithChecks(
		health.Config{
			Name:    "postgres",
			Timeout: time.Second * 5,
			Check:   healthpgx.New(healthpgx.Config{DSN: repo.GetDBURL()}),
		},
	))

	a.gEngine.GET("/health", func(c *gin.Context) {
		h.HandlerFunc(c.Writer, c.Request)
	})
}

func (a *App) Run() {
	if err := a.gEngine.Run(":" + common.Config.Port); err != nil {
		utils.LogFatal("gin.Run", slog.Any("err", err))
	}
}

func (a *App) Shutdown() {
	a.repo.Close()
	sentry.Flush(2 * time.Second)
}
