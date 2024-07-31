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

func (e *EventsAPI) GetParticipantById(c *gin.Context) {
	id := c.Param("id")
	if !e.isUserOrHasAnyRole(c, id, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u, err := e.repo.GetParticipantById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetParticipantById: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetParticipantByKeycloakID(c *gin.Context) {
	id := c.Param("id")
	if !e.isSubjectOrHasAnyRole(c, id, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u, err := e.repo.GetParticipantByKeycloakID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetParticipantByKeycloakID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetParticipantByEmail(c *gin.Context) {
	email := c.Param("email")
	if !e.isEmailOwnerOrHasAnyRole(c, email, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u, err := e.repo.GetParticipantByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetParticipantByEmail: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllParticipant(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	skip := c.Query("skip")
	limit := c.Query("limit")
	eventId := c.Query("event_id")
	eventSlug := c.Query("event_slug")
	var intEventId int

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

	if eventId != "" {
		intEventId, err = strconv.Atoi(eventId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_id value! Accepted value is INTEGER", "success": false})
			return
		}
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	u, err := e.repo.GetAllParticipants(c.Request.Context(), intSkip, intLimit, intEventId, eventSlug)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllParticipants: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) CreateNewParticipant(c *gin.Context) {
	type partWithID struct {
		repo.Part
		ID int `json:"id"`
	}

	s := repo.Part{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	if !e.isEmailOwnerOrHasAnyRole(c, *s.Email, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id, err := e.repo.CreateNewParticipant(c.Request.Context(), s)
	if err != nil {
		if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.CreateNewParticipant: %w", err))
		}
		return
	}

	partInfo := partWithID{
		Part: s,
		ID:   id,
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new participant!", "data": partInfo, "success": true})
}

func (e *EventsAPI) UpdateParticipantByID(c *gin.Context) {
	u := repo.Part{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	id := c.Param("id")
	if !e.isUserOrHasAnyRole(c, id, common.RoleRoot, common.RoleAdmin) {
		return
	}
	if err := e.repo.UpdateParticipantByID(c.Request.Context(), u, id); err != nil {
		if errors.Is(err, common.ErrNoRowsAffected) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.UpdateParticipantByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteParticipantByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id := c.Param("id")

	if err := e.repo.DeleteParticipantByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteParticipantByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant deleted successfully!", "success": true})
}
