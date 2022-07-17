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
	ID                   *int                    `json:"id" db:"id"`
	RegistrationRequired *bool                   `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string                 `json:"registration_status" db:"registration_status"`
	Audience             *string                 `json:"audience" db:"audience"`
	Slug                 *string                 `json:"slug" db:"slug"`
	Name                 *string                 `json:"name" db:"name"`
	Logo                 *string                 `json:"logo,omitempty" db:"logo"`
	Content              *map[string]interface{} `json:"content,omitempty" db:"content"`
	Deleted              *bool                   `json:"deleted" db:"deleted"`
	StartsOn             *time.Time              `json:"starts_on" db:"starts_on"`
	EndsOn               *time.Time              `json:"ends_on" db:"ends_on"`
	DateConfirmed        *bool                   `json:"date_confirmed" db:"date_confirmed"`
	ArchiveLink          *string                 `json:"archive_link" db:"archive_link"`
	CreatedAt            *time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            *time.Time              `json:"updated_at" db:"updated_at"`
	IsUserRegistered     *bool                   `json:"is_user_registered,omitempty"`
}

type event struct {
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status"`
	Audience             *string    `json:"audience" db:"audience"`
	Slug                 *string    `json:"slug" db:"slug" validate:"required"`
	Name                 *string    `json:"name" db:"name" validate:"required"`
	Logo                 *string    `json:"logo,omitempty" db:"logo"`
	Content              *string    `json:"content,omitempty" db:"content"`
	Deleted              *bool      `json:"deleted" db:"deleted"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on" validate:"required"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on" validate:"required"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed"`
	ArchiveLink          *string    `json:"archive_link" db:"archive_link"`
}

type Event interface {
	GetEventByID(ctx *gin.Context)
	GetAllEvent(ctx *gin.Context)
	CreateNewEvent(ctx *gin.Context)
	UpdateEventByID(ctx *gin.Context)
	DeleteEventByID(ctx *gin.Context)
	DeleteHardEventByID(ctx *gin.Context)
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

	u, err := getEventByID(r, ctx, id)

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
	email := ctx.Query("email")
	keycloakID := ctx.Query("kc_id")
	slug := ctx.Query("slug")

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

	fetchedEvents, err := getAllEvent(r, ctx, intSkip, intLimit, slug, email, keycloakID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": fetchedEvents, "success": true})
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

	if err := CreateEvent(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Event!", "data": s, "success": true})
}

func (r *EventDB) UpdateEventByID(ctx *gin.Context) {
	u := event{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		if err.Error() == "EOF" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "No data provided",
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

	id := ctx.Param("id")

	if err := updateEventByID(r, ctx, u, id); err != nil {

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

	if err := deleteEventByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func (r *EventDB) DeleteHardEventByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteHardEventByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func getEventByID(r *EventDB, ctx *gin.Context, id string) (eventResponse, error) {
	u := eventResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	deleted,
	starts_on,
	ends_on,
	date_confirmed,
	archive_link,
	created_at,
	updated_at 
	from event where id = $1`, id).Scan(
		&u.ID,
		&u.RegistrationRequired,
		&u.RegistrationStatus,
		&u.Audience,
		&u.Slug,
		&u.Name,
		&u.Logo,
		&u.Content,
		&u.Deleted,
		&u.StartsOn,
		&u.EndsOn,
		&u.DateConfirmed,
		&u.ArchiveLink,
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

func getAllEvent(r *EventDB, ctx *gin.Context, skip int, limit int, slug string, email string, kcID string) (*[]eventResponse, error) {

	u := []eventResponse{}

	var query string

	if email != "" || kcID != "" {

		var emailOrKcQuery string

		if email != "" {
			emailOrKcQuery = fmt.Sprintf("where p.email='%s'", email)
		} else {
			emailOrKcQuery = fmt.Sprintf("where p.keycloak_id='%s'", kcID)
		}

		query = fmt.Sprintf(`select 
			e.id,
			e.registration_required,
			e.registration_status,
			e.audience,
			e.slug,
			e.name,
			e.logo,
			e.content,
			e.deleted,
			e.starts_on,
			e.ends_on,
			e.date_confirmed,
			e.archive_link,
			e.created_at,
			e.updated_at,
			CASE WHEN (SELECT COUNT(*) FROM participation_status as ps, participant as p %s and ps.participant_id = p.id and e.id = ps.event_id ) > 0 THEN true
			ELSE false
			END AS is_user_registered 
			from event as e
			LIMIT %d OFFSET %d`, emailOrKcQuery, limit, skip)
	} else {
		whereQuery := buildAndGetWhereEventQuery(slug)

		query = fmt.Sprintf(`select 
		id,
		registration_required,
		registration_status,
		audience,
		slug,
		name,
		logo,
		content,
		deleted,
		starts_on,
		ends_on,
		date_confirmed,
		archive_link,
		created_at,
		updated_at 
		from event`+whereQuery+" LIMIT %d OFFSET %d", limit, skip)
	}

	rows, _ := r.db.Query(ctx, query)
	for rows.Next() {
		var d eventResponse
		var err error
		// Applied these checks to handle extra output is_user_registered when email or kcID is passed
		if email != "" || kcID != "" {
			err = rows.Scan(&d.ID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.Deleted, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.ArchiveLink, &d.CreatedAt, &d.UpdatedAt, &d.IsUserRegistered)
		} else {
			err = rows.Scan(&d.ID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.Deleted, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.ArchiveLink, &d.CreatedAt, &d.UpdatedAt)
		}
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateEventByID(r *EventDB, ctx *gin.Context, req event, id string) error {

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

func CreateEvent(r *EventDB, ctx *gin.Context, req event) error {

	createString, numString, createQueryArgs := prepareEventCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO event (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating event: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteEventByID(r *EventDB, ctx context.Context, id string) error {

	eventQuery := "UPDATE event SET deleted = true WHERE id=" + id + ";"
	eventItemQuery := "UPDATE event_item SET deleted = true WHERE event_id=" + id + ";"
	eventPartQuery := "UPDATE event_participation_option SET deleted = true WHERE event_id=" + id + ";"
	eventStatusQuery := "UPDATE participation_status SET deleted = true WHERE event_id=" + id + ";"
	_, err := r.db.Exec(ctx, eventQuery+eventItemQuery+eventPartQuery+eventStatusQuery)
	return err
}

func deleteHardEventByID(r *EventDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from event where id=$1", id)
	return err
}

func prepareEventUpdateQuery(req event) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

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
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
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
	if req.ArchiveLink != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("archive_link=$%d", len(updateStrings)+1))
		args = append(args, *req.ArchiveLink)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventCreateQuery(req event) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.RegistrationRequired != nil {
		createStrings = append(createStrings, "registration_required")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		createStrings = append(createStrings, "registration_status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		createStrings = append(createStrings, "audience")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		createStrings = append(createStrings, "slug")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil && *req.Logo != "" {
		createStrings = append(createStrings, "logo")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil && *req.Content != "" {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.Deleted != nil {
		createStrings = append(createStrings, "deleted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Deleted)
	}
	if req.StartsOn != nil {
		createStrings = append(createStrings, "starts_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		createStrings = append(createStrings, "ends_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		createStrings = append(createStrings, "date_confirmed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.DateConfirmed)
	}
	if req.ArchiveLink != nil {
		createStrings = append(createStrings, "archive_link")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ArchiveLink)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func buildAndGetWhereEventQuery(slug string) string {

	var whereString strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if slug != "" {
		whereCondition.WriteString(fmt.Sprintf(" slug='%s'", slug))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String()
}
