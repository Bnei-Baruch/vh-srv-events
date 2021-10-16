package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"vh-srv-event/audience"
	part "vh-srv-event/participant"
	"vh-srv-event/platform"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type Controllers struct {
	Participant         part.Participant
	ParticipationOption part.ParticipationOption
	Platform            platform.Platform
	Audience            audience.Audience
}

// cfg is the struct type that contains fields that stores the necessary configuration
// gathered from the environment.
var cfg struct {
	DBUser   string `envconfig:"DB_USER" default:"postgres"`
	DBPass   string `envconfig:"DB_PASSWORD" default:"password"`
	DBName   string `envconfig:"DB_DATABASE" default:"event"`
	DBHost   string `envconfig:"DB_HOST" default:"localhost"`
	DBPort   string `envconfig:"DB_PORT" default:"5432"`
	APP_PORT string `envconfig:"APP_PORT" default:"8080"`
}

type Router struct {
	server              *gin.Engine
	participant         part.Participant
	participationOption part.ParticipationOption
	platform            platform.Platform
	audience            audience.Audience
}

func NewRouter(server *gin.Engine, controller Controllers) *Router {
	return &Router{
		server,
		controller.Participant,
		controller.ParticipationOption,
		controller.Platform,
		controller.Audience,
	}
}
func (r *Router) Init() {

	basePath := r.server.Group("/v1")

	participant := basePath.Group("/participant")
	{
		participant.POST("/", r.participant.CreateNewParticipant)
		participant.PATCH("/:id", r.participant.UpdateParticipantByID)
		participant.DELETE("/:id", r.participant.DeleteParticipantByID)
		participant.GET("/:id", r.participant.GetParticipantById)
	}
	basePath.GET("/participants", r.participant.GetAllParticipant)

	participationOption := basePath.Group("/participation-option")
	{
		participationOption.POST("/", r.participationOption.CreateNewParticipationOption)
		participationOption.PATCH("/:name", r.participationOption.UpdateParticipationOptionByName)
		participationOption.DELETE("/:name", r.participationOption.DeleteParticipationOptionByName)
		participationOption.GET("/:name", r.participationOption.GetParticipationOptionByName)
	}
	basePath.GET("/participation-options", r.participationOption.GetAllParticipationOption)

	platform := basePath.Group("/platform")
	{
		platform.POST("/", r.platform.CreateNewPlatform)
		platform.PATCH("/:name", r.platform.UpdatePlatformByName)
		platform.DELETE("/:name", r.platform.DeletePlatformByName)
		platform.GET("/:name", r.platform.GetPlatformByName)
	}
	basePath.GET("/platforms", r.platform.GetAllPlatform)

	audience := basePath.Group("/audience")
	{
		audience.POST("/", r.audience.CreateNewAudience)
		audience.PATCH("/:name", r.audience.UpdateAudienceByName)
		audience.DELETE("/:name", r.audience.DeleteAudienceByName)
		audience.GET("/:name", r.audience.GetAudienceByName)
	}
	basePath.GET("/audiences", r.audience.GetAllAudience)
}

func main() {
	route := gin.Default()

	if err := envconfig.Process("LIST", &cfg); err != nil {
		log.Fatalln("Error while fetching env file")
		return
	}

	databaseURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPass + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	participant := part.NewParticipant(conn)
	participationOption := part.NewParticipationOption(conn)
	platform := platform.NewPlatform(conn)
	audience := audience.NewAudience(conn)

	r := NewRouter(route, Controllers{
		Participant:         participant,
		ParticipationOption: participationOption,
		Platform:            platform,
		Audience:            audience,
	})

	r.Init()

	route.Run("localhost:" + cfg.APP_PORT)
}
