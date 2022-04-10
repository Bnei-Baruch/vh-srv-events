package notification

import (
	"fmt"
	"net/http"
	"vh-srv-event/participant"
	"vh-srv-event/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4/pgxpool"
)

type notification struct {
	Language      *string `json:"language" validate:"required"`
	EventSlug     *string `json:"event_slug" validate:"required"`
	EventTemplate *string `json:"email_template" validate:"required"`
	FromEmail     *string `json:"from_email"`
	FromName      *string `json:"from_name"`
}

type Notification interface {
	SendEventEmail(ctx *gin.Context)
}

type NotificationDB struct {
	db *pgxpool.Pool
}

func NewNotification(db *pgxpool.Pool) Notification {
	return &NotificationDB{
		db,
	}
}

func (r *NotificationDB) SendEventEmail(ctx *gin.Context) {

	s := notification{}
	if err := ctx.ShouldBindJSON(&s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	err := validator.New().Struct(s)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	userFetchErr := fetchUsersAndSendEmail(r, ctx, s)

	if userFetchErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   userFetchErr.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": "", "success": true})
}

func fetchUsersAndSendEmail(r *NotificationDB, ctx *gin.Context, s notification) error {
	rows, err := r.db.Query(ctx, fmt.Sprintf(
		`SELECT DISTINCT p.email, p.first_name, p.last_name from participant as p, participation_status as ps, event as e
		WHERE ps.event_id = (SELECT id FROM event WHERE slug = '%s') AND
		ps.participant_id = p.id AND
		p.email_language = '%s';
	`, *s.EventSlug, *s.Language))
	if err != nil {
		return err
	}
	for rows.Next() {
		var d participant.Part
		err := rows.Scan(&d.Email, &d.FirstName, &d.LastName)
		if err != nil {
			return err
		}
		emailErr := util.SendEmail(s.FromName, s.FromEmail, *s.EventTemplate, *s.Language, *d.Email, *d.FirstName, *d.LastName)
		if emailErr != nil {
			return emailErr
		}
	}
	return nil
}
