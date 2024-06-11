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

func (e *EventsAPI) GetEventPartOptionByID(c *gin.Context) {
	id := c.Param("id")

	u, err := e.repo.GetEventPartOptionByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetEventPartOptionByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllEventPartOption(c *gin.Context) {
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

	u, err := e.repo.GetAllEventPartOption(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllEventPartOption: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) CreateNewEventPartOption(c *gin.Context) {
	s := repo.EventPartOption{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if err := e.repo.CreateNewEventPartOption(c.Request.Context(), s); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateNewEventPartOption: %w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new Event Participation Option!", "data": s, "success": true})
}

func (e *EventsAPI) UpdateEventPartOptionByID(c *gin.Context) {
	u := repo.EventPartOption{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	id := c.Param("id")
	if err := e.repo.UpdateEventPartOptionByID(c.Request.Context(), u, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.UpdateEventPartOptionByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event Participation Option updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteEventPartOptionByID(c *gin.Context) {
	id := c.Param("id")

	if err := e.repo.DeleteEventPartOptionByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteEventPartOptionByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event Participation Option deleted successfully!", "success": true})
}
