package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

type PartAnalyticRes struct {
	PartOptionAndCount []repo.PartOptionAndCount `json:"part_option_details"`
	TotalParticpant    int                       `json:"total_participant"`
}

func (e *EventsAPI) PartAnalytics(c *gin.Context) {
	eventId, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	partOptAndCount, partOptionsAndCountErr := e.repo.FetchTotalParticipantByOptionAndGroupBy(c.Request.Context(), uint(eventId))
	if partOptionsAndCountErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": partOptionsAndCountErr.Error(), "success": false})
		return
	}

	totalPartOption, totalPartOptionErr := e.repo.FetchTotalParticipantByOption(c.Request.Context(), uint(eventId))
	if totalPartOptionErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": totalPartOptionErr.Error(), "success": false})
		return
	}

	res := PartAnalyticRes{
		PartOptionAndCount: partOptAndCount,
		TotalParticpant:    totalPartOption,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": res, "success": true})
}
