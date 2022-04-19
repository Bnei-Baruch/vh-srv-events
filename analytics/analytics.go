package analytics

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

type partOptionAndCount struct {
	ParticipationOption *string `json:"participation_option"`
	Count               *int    `json:"count"`
}

type partAnalyticRes struct {
	PartOptionAndCount []partOptionAndCount `json:"part_option_details"`
	TotalParticpant    int                  `json:"total_participant"`
}

type Analytics interface {
	PartAnalytics(ctx *gin.Context)
}

type AnalyticsDB struct {
	db *pgxpool.Pool
}

func NewAnalytics(db *pgxpool.Pool) Analytics {
	return &AnalyticsDB{
		db,
	}
}

func (r *AnalyticsDB) PartAnalytics(ctx *gin.Context) {

	eventId := ctx.Query("event_id")

	partOptAndCount, partOptionsAndCountErr := fetchTotalParticipantByOptionAndGroupBy(r, ctx, eventId)

	if partOptionsAndCountErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   partOptionsAndCountErr.Error(),
			"success": false,
		})
		return
	}

	totalPartOption, totalPartOptionErr := fetchTotalParticipantByOption(r, ctx, eventId)

	if totalPartOptionErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   totalPartOptionErr.Error(),
			"success": false,
		})
		return
	}

	res := partAnalyticRes{
		PartOptionAndCount: partOptAndCount,
		TotalParticpant:    totalPartOption,
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": res, "success": true})
}

func fetchTotalParticipantByOptionAndGroupBy(r *AnalyticsDB, ctx *gin.Context, eventId string) ([]partOptionAndCount, error) {

	partOptions := []partOptionAndCount{}
	var eventIDQuery string
	if eventId != "" {
		eventIDQuery = ` WHERE event_id=` + eventId
	}
	rows, err := r.db.Query(ctx, `select participation_option, count (participation_option) as qt 
		from participation_status`+eventIDQuery+
		` group by participation_option`)
	if err != nil {
		return partOptions, err
	}
	for rows.Next() {
		var p partOptionAndCount
		err := rows.Scan(&p.ParticipationOption, &p.Count)
		if err != nil {
			return partOptions, err
		}
		partOptions = append(partOptions, p)
	}
	return partOptions, nil
}
func fetchTotalParticipantByOption(r *AnalyticsDB, ctx *gin.Context, eventId string) (int, error) {

	var count int
	var eventIDQuery string
	if eventId != "" {
		eventIDQuery = ` WHERE event_id=` + eventId
	}
	if err := r.db.QueryRow(ctx, `select count (participation_option) as qt from participation_status`+eventIDQuery).Scan(
		&count,
	); err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("not found")
		}
		return 0, err
	}
	return count, nil
}
