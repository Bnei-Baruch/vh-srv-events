package audience

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type audienceResponse struct {
	Name *string `json:"name" db:"name"`
}

type audience struct {
	Name *string `json:"Name" db:"Name" validate:"required"`
}

type Audience interface {
	GetAudienceByName(ctx *gin.Context)
	GetAllAudience(ctx *gin.Context)
	CreateNewAudience(ctx *gin.Context)
	UpdateAudienceByName(ctx *gin.Context)
	DeleteAudienceByName(ctx *gin.Context)
}

type AudienceDB struct {
	db *pgxpool.Pool
}

func NewAudience(db *pgxpool.Pool) Audience {
	return &AudienceDB{
		db,
	}
}

func (r *AudienceDB) GetAudienceByName(ctx *gin.Context) {
	name := ctx.Param("name")

	u, err := getAudienceByName(r, ctx, name)

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

func (r *AudienceDB) GetAllAudience(ctx *gin.Context) {
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

	u, err := GetAllAudience(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *AudienceDB) CreateNewAudience(ctx *gin.Context) {
	s := audience{}
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

	if err := CreateNewAudience(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Created new audience!", "data": s, "success": true})
}

func (r *AudienceDB) UpdateAudienceByName(ctx *gin.Context) {
	u := audience{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	name := ctx.Param("name")

	if err := UpdateAudienceByName(r, ctx, u, name); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Audience updated successfully", "data": u, "success": true})
}

func (r *AudienceDB) DeleteAudienceByName(ctx *gin.Context) {

	name := ctx.Param("name")

	if err := DeleteAudienceByName(r, ctx, name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Audience deleted successfully!", "success": true})
}

func getAudienceByName(r *AudienceDB, ctx *gin.Context, name string) (audienceResponse, error) {
	u := audienceResponse{}
	if err := r.db.QueryRow(ctx, `select 
	name 
	from audience where name = $1`, name).Scan(
		&u.Name,
	); err != nil {
		if err == pgx.ErrNoRows {
			return audienceResponse{}, fmt.Errorf("not found")
		}
		return audienceResponse{}, err
	}
	return u, nil
}

func GetAllAudience(r *AudienceDB, ctx *gin.Context, skip int, limit int) (*[]audienceResponse, error) {

	u := []audienceResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	name 
	from audience LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d audienceResponse
		err := rows.Scan(&d.Name)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdateAudienceByName(r *AudienceDB, ctx *gin.Context, req audience, name string) error {
	if req.Name != nil {
		updateRes, err := r.db.Exec(ctx, `UPDATE audience SET name=$1 WHERE name=$2`, req.Name, name)
		if err != nil {
			return fmt.Errorf("problem updating audience: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func CreateNewAudience(r *AudienceDB, ctx *gin.Context, req audience) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO audience (
			name)
		VALUES (
			$1)  `,
		*req.Name)

	return err
}

func DeleteAudienceByName(r *AudienceDB, ctx context.Context, name string) error {
	_, err := r.db.Exec(ctx, "delete from audience where name=$1", name)
	return err
}
