package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

func (e *EventsAPI) SendEventEmail(c *gin.Context) {
	s := repo.Notification{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err = e.repo.FetchUsersAndSendEmail(c.Request.Context(), s)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.FetchUsersAndSendEmail: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": "", "success": true})
}
