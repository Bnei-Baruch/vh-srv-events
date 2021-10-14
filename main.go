package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ctrl "vh-srv-event/controller"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type Controllers struct {
	Participant ctrl.Participant
}

// cfg is the struct type that contains fields that stores the necessary configuration
// gathered from the environment.
var cfg struct {
	DBUser string `envconfig:"DB_USER" default:"postgres"`
	DBPass string `envconfig:"DB_PASSWORD" default:"password"`
	DBName string `envconfig:"DB_DATABASE" default:"event"`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort string `envconfig:"DB_PORT" default:"5432"`
}

type Router struct {
	server      *gin.Engine
	participant ctrl.Participant
}

func NewRouter(server *gin.Engine, controller Controllers) *Router {
	return &Router{
		server,
		controller.Participant,
	}
}
func (r *Router) Init() {

	basePath := r.server.Group("/v1")

	basePath.GET("/health", ctrl.Health)

	participant := basePath.Group("/participant")
	{
		participant.POST("/", r.participant.CreateNewParticipant)
		participant.PATCH("/:id", r.participant.UpdateParticipantByID)
		participant.DELETE("/:id", r.participant.DeleteParticipantByID)
		participant.GET("/:id", r.participant.GetParticipantById)
	}
	basePath.GET("/participants", r.participant.GetAllParticipant)
}

func main() {
	route := gin.Default()

	if err := envconfig.Process("LIST", &cfg); err != nil {
		log.Fatalln("Error while fetching env file")
		return
	}

	databaseURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPass + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName

	conn, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	participant := ctrl.NewParticipant(conn)

	r := NewRouter(route, Controllers{
		Participant: participant,
	})

	r.Init()

	route.Run("localhost:8080")
}
