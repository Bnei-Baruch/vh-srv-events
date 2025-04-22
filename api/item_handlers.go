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

func (e *EventsAPI) GetItemByID(c *gin.Context) {
	id := c.Param("id")
	u, err := e.repo.GetItemByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetItemByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllItem(c *gin.Context) {
	skip := c.Query("skip")
	limit := c.Query("limit")

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "100"
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

	u, err := e.repo.GetAllItem(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllItem: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) CreateNewItem(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	s := repo.Item{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if err := e.repo.CreateNewItem(c.Request.Context(), s); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateNewItem: %w", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new Item!", "data": s, "success": true})
}

func (e *EventsAPI) UpdateItemByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u := repo.Item{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	id := c.Param("id")
	if err := e.repo.UpdateItemByID(c.Request.Context(), u, id); err != nil {
		if errors.Is(err, common.ErrNoRowsAffected) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.UpdateItemByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteItemByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id := c.Param("id")

	if err := e.repo.DeleteItemByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteItemByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully!", "success": true})
}
