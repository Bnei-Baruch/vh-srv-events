package participant

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

type partResponse struct {
	ID            *int       `json:"id" db:"id"`
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	FirstLanguage *string    `json:"first_language,omitempty" db:"first_language"`
	EmailLanguage *string    `json:"email_language,omitempty" db:"email_language"`
	DOB           *time.Time `json:"dob,omitempty" db:"dob"`
	Gender        *string    `json:"gender,omitempty" db:"gender"`
	Email         *string    `json:"email" db:"email"`
	Country       *string    `json:"country,omitempty" db:"country"`
	PhoneNumber   *string    `json:"phone_number,omitempty" db:"phone_number"`
	FirstName     *string    `json:"first_name" db:"first_name"`
	LastName      *string    `json:"last_name" db:"last_name"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

type Part struct {
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id" validate:"required,uuid"`
	FirstLanguage *string    `json:"first_language,omitempty" db:"first_language"`
	EmailLanguage *string    `json:"email_language,omitempty" db:"email_language"`
	DOB           *time.Time `json:"dob,omitempty" db:"dob"`
	Gender        *string    `json:"gender,omitempty" db:"gender"`
	Email         *string    `json:"email" db:"email" validate:"required,email"`
	Country       *string    `json:"country,omitempty" db:"country"`
	PhoneNumber   *string    `json:"phone_number,omitempty" db:"phone_number"`
	FirstName     *string    `json:"first_name" db:"first_name" validate:"required"`
	LastName      *string    `json:"last_name" db:"last_name" validate:"required"`
}

type Participant interface {
	GetParticipantById(ctx *gin.Context)
	GetParticipantByEmail(ctx *gin.Context)
	GetParticipantByKeycloakID(ctx *gin.Context)
	GetAllParticipant(ctx *gin.Context)
	CreateNewParticipant(ctx *gin.Context)
	UpdateParticipantByID(ctx *gin.Context)
	DeleteParticipantByID(ctx *gin.Context)
}

type ParticipantDB struct {
	db *pgxpool.Pool
}

func NewParticipant(db *pgxpool.Pool) Participant {
	return &ParticipantDB{
		db,
	}
}

func (r *ParticipantDB) GetParticipantById(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getPartById(r, ctx, id)

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

func (r *ParticipantDB) GetParticipantByKeycloakID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getPartByKeycloakID(r, ctx, id)

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

func (r *ParticipantDB) GetParticipantByEmail(ctx *gin.Context) {
	email := ctx.Param("email")

	u, err := getPartByEmail(r, ctx, email)

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

func (r *ParticipantDB) GetAllParticipant(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")
	eventId := ctx.Query("event_id")
	eventSlug := ctx.Query("event_slug")
	var intEventId int

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
	if eventId != "" {
		intEventId, err = strconv.Atoi(eventId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id value! Accepted value is INTEGER", "success": false})
			return
		}
	}

	// String conversion to int
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	u, err := GetAllPart(r, ctx, intSkip, intLimit, intEventId, eventSlug)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ParticipantDB) CreateNewParticipant(ctx *gin.Context) {
	type partWithID struct {
		Part
		ID int `json:"id"`
	}
	s := Part{}
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

	id, err := CreateNewPart(r, ctx, s)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	partInfo := partWithID{
		Part: s,
		ID:   id,
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new participant!", "data": partInfo, "success": true})
}

func (r *ParticipantDB) UpdateParticipantByID(ctx *gin.Context) {
	u := Part{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := UpdatePartByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Participant updated successfully", "data": u, "success": true})
}

func (r *ParticipantDB) DeleteParticipantByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := DeletePartByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participant deleted successfully!", "success": true})
}

func getPartById(r *ParticipantDB, ctx *gin.Context, id string) (partResponse, error) {
	u := partResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where id = $1`, id).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return partResponse{}, fmt.Errorf("not found")
		}
		return partResponse{}, err
	}
	return u, nil
}

func getPartByEmail(r *ParticipantDB, ctx *gin.Context, email string) (partResponse, error) {
	u := partResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where email = $1`, email).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return partResponse{}, fmt.Errorf("not found")
		}
		return partResponse{}, err
	}
	return u, nil
}

func getPartByKeycloakID(r *ParticipantDB, ctx *gin.Context, id string) (partResponse, error) {
	u := partResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where keycloak_id = $1`, id).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return partResponse{}, fmt.Errorf("not found")
		}
		return partResponse{}, err
	}
	return u, nil
}

func GetAllPart(r *ParticipantDB, ctx *gin.Context, skip int, limit int, eventId int, eventSlug string) (*[]partResponse, error) {

	u := []partResponse{}

	var query string

	if eventId != 0 || eventSlug != "" {
		var eventlOrSlugWhereQuery string
		var eventlOrSlugFromQuery string

		if eventId != 0 {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps`
			eventlOrSlugWhereQuery = fmt.Sprintf(`WHERE ps.event_id = %d 
			AND p.id = ps.participant_id`, eventId)
		} else {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps, event as e`
			eventlOrSlugWhereQuery = fmt.Sprintf(`WHERE e.slug = '%s'
			AND ps.event_id = e.id
			AND p.id = ps.participant_id`, eventSlug)
		}

		query = fmt.Sprintf(`SELECT 
			p.id,
			p.keycloak_id,
			p.first_language,
			p.email_language,
			p.dob,
			p.gender,
			p.email,
			p.country,
			p.first_name,
			p.last_name,
			p.created_at,
			p.updated_at 
			%s
			%s
			LIMIT %d OFFSET %d`, eventlOrSlugFromQuery, eventlOrSlugWhereQuery, limit, skip)

	} else {
		query = fmt.Sprintf(`SELECT 
			id,
			keycloak_id,
			first_language,
			email_language,
			dob,
			gender,
			email,
			country,
			first_name,
			last_name,
			created_at,
			updated_at 
			FROM participant LIMIT %d OFFSET %d`, limit, skip)
	}

	rows, _ := r.db.Query(ctx, query)
	for rows.Next() {
		var d partResponse
		err := rows.Scan(&d.ID, &d.KeycloakID, &d.FirstLanguage, &d.EmailLanguage, &d.DOB, &d.Gender, &d.Email, &d.Country, &d.FirstName, &d.LastName, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdatePartByID(r *ParticipantDB, ctx *gin.Context, req Part, id string) error {
	toUpdate, toUpdateArgs := prepareParticipantUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE participant SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating participant: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func CreateNewPart(r *ParticipantDB, ctx *gin.Context, req Part) (int, error) {

	createString, numString, createQueryArgs := prepareParticipantCreateQuery(req)

	var ID int

	if len(createQueryArgs) != 0 {
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO participant (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(
			&ID,
		); err != nil {
			if err == pgx.ErrNoRows {
				return 0, fmt.Errorf("no rows affected")
			}
			return 0, err
		}
		return ID, nil
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func DeletePartByID(r *ParticipantDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from participant where id=$1", id)
	return err
}

func prepareParticipantUpdateQuery(req Part) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.KeycloakID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("keycloak_id=$%d", len(updateStrings)+1))
		args = append(args, *req.KeycloakID)
	}
	if req.FirstLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_language=$%d", len(updateStrings)+1))
		args = append(args, *req.FirstLanguage)
	}
	if req.EmailLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email_language=$%d", len(updateStrings)+1))
		args = append(args, *req.EmailLanguage)
	}
	if req.DOB != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("dob=$%d", len(updateStrings)+1))
		args = append(args, *req.DOB)
	}
	if req.Gender != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("gender=$%d", len(updateStrings)+1))
		args = append(args, *req.Gender)
	}
	if req.Email != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email=$%d", len(updateStrings)+1))
		args = append(args, *req.Email)
	}
	if req.Country != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("country=$%d", len(updateStrings)+1))
		args = append(args, *req.Country)
	}
	if req.PhoneNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("phone_number=$%d", len(updateStrings)+1))
		args = append(args, *req.PhoneNumber)
	}
	if req.FirstName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name=$%d", len(updateStrings)+1))
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name=$%d", len(updateStrings)+1))
		args = append(args, *req.LastName)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareParticipantCreateQuery(req Part) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.KeycloakID != nil {
		createStrings = append(createStrings, "keycloak_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.KeycloakID)
	}
	if req.FirstLanguage != nil {
		createStrings = append(createStrings, "first_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.FirstLanguage)
	}
	if req.EmailLanguage != nil {
		createStrings = append(createStrings, "email_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EmailLanguage)
	}
	if req.DOB != nil {
		createStrings = append(createStrings, "dob")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.DOB)
	}
	if req.Gender != nil {
		createStrings = append(createStrings, "gender")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Gender)
	}
	if req.Email != nil {
		createStrings = append(createStrings, "email")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Email)
	}
	if req.Country != nil {
		createStrings = append(createStrings, "country")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Country)
	}
	if req.PhoneNumber != nil {
		createStrings = append(createStrings, "phone_number")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.PhoneNumber)
	}
	if req.FirstName != nil {
		createStrings = append(createStrings, "first_name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		createStrings = append(createStrings, "last_name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.LastName)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
