package event

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type eventResponse struct {
	ID                   *int       `json:"id" db:"id"`
	ParticipantID        *int       `json:"participant_id" db:"participant_id"`
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status"`
	Audience             *string    `json:"audience" db:"audience"`
	Slug                 *string    `json:"slug" db:"slug"`
	Name                 *string    `json:"name" db:"name"`
	Logo                 *string    `json:"logo" db:"logo"`
	Content              *string    `json:"content" db:"content"`
	ItemID               *int       `json:"item_id" db:"item_id"`
	ParticipationOption  *int       `json:"participation_option" db:"participation_option"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed"`
	CreatedAt            *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at" db:"updated_at"`
}

type event struct {
	ParticipantID        *int       `json:"participant_id" db:"participant_id" validate:"required"`
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required" validate:"required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status" validate:"required"`
	Audience             *string    `json:"audience" db:"audience" validate:"required"`
	Slug                 *string    `json:"slug" db:"slug" validate:"required"`
	Name                 *string    `json:"name" db:"name" validate:"required"`
	Logo                 *string    `json:"logo" db:"logo" validate:"required"`
	Content              *string    `json:"content" db:"content" validate:"required"`
	ItemID               *int       `json:"item_id" db:"item_id" validate:"required"`
	ParticipationOption  *int       `json:"participation_option" db:"participation_option" validate:"required"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on" validate:"required"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on" validate:"required"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed" validate:"required"`
}

type Event interface {
	GetEventByID(ctx *gin.Context)
	GetAllEvent(ctx *gin.Context)
	CreateNewEvent(ctx *gin.Context)
	UpdateEventByID(ctx *gin.Context)
	DeleteEventByID(ctx *gin.Context)
}

type EventDB struct {
	db *pgxpool.Pool
}

func NewEvent(db *pgxpool.Pool) Event {
	return &EventDB{
		db,
	}
}

func (r *EventDB) GetEventByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getURLByID(r, ctx, id)

	if err != nil {
		if err.Error() == "not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *EventDB) GetAllEvent(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "10"
	}

	// String conversion to int
	intSkip, err := strconv.Atoi(skip)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER", "success": false})
		return
	}

	// String conversion to int
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	u, err := getAllURL(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *EventDB) CreateNewEvent(ctx *gin.Context) {
	s := event{}
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

	if err := createNewURL(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Created new Event!", "data": s, "success": true})
}

func (r *EventDB) UpdateEventByID(ctx *gin.Context) {
	u := event{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateURLByID(r, ctx, u, id); err != nil {

		if err.Error() == "not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		if err.Error() == "invalid values" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Event updated successfully", "data": u, "success": true})
}

func (r *EventDB) DeleteEventByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteURLByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func getURLByID(r *EventDB, ctx *gin.Context, id string) (eventResponse, error) {
	u := eventResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	participant_id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	item_id,
	participation_option,
	starts_on,
	ends_on,
	date_confirmed,
	created_at,
	updated_at 
	from event where id = $1`, id).Scan(
		&u.ID,
		&u.ParticipantID,
		&u.RegistrationRequired,
		&u.RegistrationStatus,
		&u.Audience,
		&u.Slug,
		&u.Name,
		&u.Logo,
		&u.Content,
		&u.ItemID,
		&u.ParticipationOption,
		&u.StartsOn,
		&u.EndsOn,
		&u.DateConfirmed,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return eventResponse{}, fmt.Errorf("not found")
		}
		return eventResponse{}, err
	}
	return u, nil
}

func getAllURL(r *EventDB, ctx *gin.Context, skip int, limit int) (*[]eventResponse, error) {

	u := []eventResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	participant_id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	item_id,
	participation_option,
	starts_on,
	ends_on,
	date_confirmed,
	created_at,
	updated_at 
	from event LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d eventResponse
		err := rows.Scan(&d.ID, &d.ParticipantID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.ItemID, &d.ParticipationOption, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateURLByID(r *EventDB, ctx *gin.Context, req event, id string) error {

	toUpdate, toUpdateArgs := prepareEventUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE event SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating event: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewURL(r *EventDB, ctx *gin.Context, req event) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO event (
			participant_id,
			registration_required,
			registration_status,
			audience,
			slug,
			name,
			logo,
			content,
			item_id,
			participation_option,
			starts_on,
			ends_on,
			date_confirmed)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13)  `,
		*req.ParticipantID,
		*req.RegistrationRequired,
		*req.RegistrationStatus,
		*req.Audience,
		*req.Slug,
		*req.Name,
		*req.Logo,
		*req.Content,
		*req.ItemID,
		*req.ParticipationOption,
		*req.StartsOn,
		*req.EndsOn,
		*req.DateConfirmed)

	return err
}

func deleteURLByID(r *EventDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from event where id=$1", id)
	return err
}

func prepareEventUpdateQuery(req event) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.ParticipantID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participant_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipantID)
	}
	if req.RegistrationRequired != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_required=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_status=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("audience=$%d", len(updateStrings)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("slug=$%d", len(updateStrings)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("logo=$%d", len(updateStrings)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}
	if req.ItemID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("item_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ItemID)
	}
	if req.ParticipationOption != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participation_option=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.StartsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("starts_on=$%d", len(updateStrings)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("ends_on=$%d", len(updateStrings)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("date_confirmed=$%d", len(updateStrings)+1))
		args = append(args, *req.DateConfirmed)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
