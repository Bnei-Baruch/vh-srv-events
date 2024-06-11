package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

func (e *EventsAPI) GetParticipationOptionByName(c *gin.Context) {
	name := c.Param("name")

	u, err := e.repo.GetParticipantOptionByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetParticipantOptionByName: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllParticipationOption(c *gin.Context) {
	skip := c.Query("skip")
	limit := c.Query("limit")

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "10"
	}

	intSkip, err := strconv.Atoi(skip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER", "success": false})
		return
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	u, err := e.repo.GetAllParticipantOption(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllParticipantOption: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) CreateNewParticipationOption(c *gin.Context) {
	s := repo.ParticipantOption{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if err := e.repo.CreateNewParticipantOption(c.Request.Context(), s); err != nil {
		if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.CreateNewParticipantOption: %w", err))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new participation option!", "data": s, "success": true})
}

func (e *EventsAPI) UpdateParticipationOptionByName(c *gin.Context) {
	u := repo.ParticipantOption{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	name := c.Param("name")

	if err := e.repo.UpdateParticipantOptionByName(c.Request.Context(), u, name); err != nil {
		if errors.Is(err, common.ErrNoRowsAffected) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.UpdateParticipantOptionByName: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participation option updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteParticipationOptionByName(c *gin.Context) {
	name := c.Param("name")

	if err := e.repo.DeleteParticipantOptionByName(c.Request.Context(), name); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteParticipantOptionByName: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participation option deleted successfully!", "success": true})
}
