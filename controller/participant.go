package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

type partStruct struct {
	ID            *int       `json:"id" db:"id"`
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	FirstLanguage *string    `json:"first_language" db:"first_language"`
	EmailLanguage *string    `json:"email_language" db:"email_language"`
	DOB           *time.Time `json:"dob" db:"dob"`
	Gender        *string    `json:"gender" db:"gender"`
	Email         *string    `json:"email" db:"email"`
	Country       *string    `json:"country" db:"country"`
	FirstName     *string    `json:"first_name" db:"first_name"`
	LastName      *string    `json:"last_name" db:"last_name"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

type partRequest struct {
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	FirstLanguage *string    `json:"first_language" db:"first_language"`
	EmailLanguage *string    `json:"email_language" db:"email_language"`
	DOB           *time.Time `json:"dob" db:"dob"`
	Gender        *string    `json:"gender" db:"gender"`
	Email         *string    `json:"email" db:"email"`
	Country       *string    `json:"country" db:"country"`
	FirstName     *string    `json:"first_name" db:"first_name"`
	LastName      *string    `json:"last_name" db:"last_name"`
}

type Participant interface {
	GetParticipantById(ctx *gin.Context)
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

func (r *ParticipantDB) GetAllParticipant(ctx *gin.Context) {
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

	u, err := GetAllPart(r, ctx, intSkip, intLimit)

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
	s := partRequest{}
	if err := ctx.ShouldBindJSON(&s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if s.KeycloakID == nil || s.FirstLanguage == nil || s.EmailLanguage == nil || s.DOB == nil || s.Gender == nil || s.Email == nil || s.Country == nil || s.FirstName == nil || s.LastName == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid body parameters",
			"success": false,
		})
		return
	}

	if err := CreateNewPart(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Created new participant!", "data": s, "success": true})
}

func (r *ParticipantDB) UpdateParticipantByID(ctx *gin.Context) {
	u := partRequest{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if u.KeycloakID == nil && u.FirstLanguage == nil && u.EmailLanguage == nil && u.DOB == nil && u.Gender == nil && u.Email == nil && u.Country == nil && u.FirstName == nil && u.LastName == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid body parameters",
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

func getPartById(r *ParticipantDB, ctx *gin.Context, id string) (partStruct, error) {
	u := partStruct{}
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
			return partStruct{}, fmt.Errorf("not found")
		}
		return partStruct{}, err
	}
	return u, nil
}

func GetAllPart(r *ParticipantDB, ctx *gin.Context, skip int, limit int) (*[]partStruct, error) {

	u := []partStruct{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
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
	from participant LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d partStruct
		err := rows.Scan(&d.ID, &d.KeycloakID, &d.FirstLanguage, &d.EmailLanguage, &d.DOB, &d.Gender, &d.Email, &d.Country, &d.FirstName, &d.LastName, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdatePartByID(r *ParticipantDB, ctx *gin.Context, req partRequest, id string) error {
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
		return fmt.Errorf("problem updating participant")
	}
}

func CreateNewPart(r *ParticipantDB, ctx *gin.Context, req partRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO participant (
			keycloak_id,
			first_language,
			email_language,
			dob,
			gender,
			email,
			country,
			first_name,
			last_name)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9)  `,
		*req.KeycloakID,
		*req.FirstLanguage,
		*req.EmailLanguage,
		*req.DOB,
		*req.Gender,
		*req.Email,
		*req.Country,
		*req.FirstName,
		*req.LastName)

	return err
}

func DeletePartByID(r *ParticipantDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from participant where id=$1", id)
	return err
}

func prepareParticipantUpdateQuery(req partRequest) (string, []interface{}) {
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
