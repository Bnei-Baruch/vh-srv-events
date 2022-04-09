package partoptn

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type partOptionResponse struct {
	Name        *string                 `json:"name" db:"name"`
	Description *string                 `json:"description,omitempty" db:"description"`
	Content     *map[string]interface{} `json:"content,omitempty" db:"content"`
}

type partOption struct {
	Name        *string `json:"name" db:"Name" validate:"required"`
	Description *string `json:"description,omitempty" db:"description"`
	Content     *string `json:"content,omitempty" db:"content"`
}

type ParticipationOption interface {
	GetParticipationOptionByName(ctx *gin.Context)
	GetAllParticipationOption(ctx *gin.Context)
	CreateNewParticipationOption(ctx *gin.Context)
	UpdateParticipationOptionByName(ctx *gin.Context)
	DeleteParticipationOptionByName(ctx *gin.Context)
}

type ParticipationOptionDB struct {
	db *pgxpool.Pool
}

func NewParticipationOption(db *pgxpool.Pool) ParticipationOption {
	return &ParticipationOptionDB{
		db,
	}
}

func (r *ParticipationOptionDB) GetParticipationOptionByName(ctx *gin.Context) {
	name := ctx.Param("name")

	u, err := getPartOptionByName(r, ctx, name)

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

func (r *ParticipationOptionDB) GetAllParticipationOption(ctx *gin.Context) {
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

	u, err := GetAllPartOption(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ParticipationOptionDB) CreateNewParticipationOption(ctx *gin.Context) {
	s := partOption{}
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

	if err := CreateNewPartOption(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new participation option!", "data": s, "success": true})
}

func (r *ParticipationOptionDB) UpdateParticipationOptionByName(ctx *gin.Context) {
	u := partOption{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	name := ctx.Param("name")

	if err := UpdatePartOptionByName(r, ctx, u, name); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Participation option updated successfully", "data": u, "success": true})
}

func (r *ParticipationOptionDB) DeleteParticipationOptionByName(ctx *gin.Context) {

	name := ctx.Param("name")

	if err := DeletePartOptionByName(r, ctx, name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participation option deleted successfully!", "success": true})
}

func getPartOptionByName(r *ParticipationOptionDB, ctx *gin.Context, name string) (partOptionResponse, error) {
	u := partOptionResponse{}
	if err := r.db.QueryRow(ctx, `select 
	name,
	description,
	content
	from participation_option where name = $1`, name).Scan(
		&u.Name,
		&u.Description,
		&u.Content,
	); err != nil {
		if err == pgx.ErrNoRows {
			return partOptionResponse{}, fmt.Errorf("not found")
		}
		return partOptionResponse{}, err
	}
	return u, nil
}

func GetAllPartOption(r *ParticipationOptionDB, ctx *gin.Context, skip int, limit int) (*[]partOptionResponse, error) {

	u := []partOptionResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	name,
	description,
	content
	from participation_option LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d partOptionResponse
		err := rows.Scan(&d.Name, &d.Description, &d.Content)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdatePartOptionByName(r *ParticipationOptionDB, ctx *gin.Context, req partOption, name string) error {

	toUpdate, toUpdateArgs := preparePartOptionUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE participation_option SET %s WHERE name='%s'`, toUpdate, name),
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

func CreateNewPartOption(r *ParticipationOptionDB, ctx *gin.Context, req partOption) error {

	createString, numString, createQueryArgs := preparePartOptionCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO participation_option (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating event: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func DeletePartOptionByName(r *ParticipationOptionDB, ctx context.Context, name string) error {
	_, err := r.db.Exec(ctx, "delete from participation_option where name=$1", name)
	return err
}

func preparePartOptionCreateQuery(req partOption) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Content != nil {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.Description != nil {
		createStrings = append(createStrings, "description")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Description)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func preparePartOptionUpdateQuery(req partOption) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("description=$%d", len(updateStrings)+1))
		args = append(args, *req.Description)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
