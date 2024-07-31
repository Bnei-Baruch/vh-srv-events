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

func (e *EventsAPI) GetEventByID(c *gin.Context) {
	id := c.Param("id")
	u, err := e.repo.GetEventByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetEventByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllEvent(c *gin.Context) {

	skip := c.Query("skip")
	limit := c.Query("limit")
	email := c.Query("email")
	keycloakID := c.Query("kc_id")
	slug := c.Query("slug")

	if email != "" {
		if !e.isEmailOwnerOrHasAnyRole(c, email, common.RoleRoot, common.RoleAdmin) {
			return
		}
	}
	if keycloakID != "" {
		if !e.isSubjectOrHasAnyRole(c, keycloakID, common.RoleRoot, common.RoleAdmin) {
			return
		}
	} else {
		if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
			var ok bool
			keycloakID, ok = e.getUserKeyFromRequest(c)
			if !ok {
				return
			}
		}
	}
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

	fetchedEvents, err := e.repo.GetAllEvent(c.Request.Context(), intSkip, intLimit, slug, email, keycloakID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllEvent: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": fetchedEvents, "success": true})
}

func (e *EventsAPI) CreateNewEvent(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	s := repo.Event{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if err := e.repo.CreateEvent(c.Request.Context(), s); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateEvent: %w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new Event!", "data": s, "success": true})
}

func (e *EventsAPI) UpdateEventByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u := repo.Event{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	id := c.Param("id")

	if err := e.repo.UpdateEventByID(c.Request.Context(), u, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.UpdateEventByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteEventByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id := c.Param("id")

	if err := e.repo.DeleteEventByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteEventByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func (e *EventsAPI) DeleteHardEventByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id := c.Param("id")

	if err := e.repo.DeleteHardEventByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteHardEventByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}
