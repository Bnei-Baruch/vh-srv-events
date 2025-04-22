package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"github.com/jackc/pgx/v4"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
	"gitlab.bbdev.team/vh/vh-srv-events/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

func (e *EventsAPI) GetParticipationStatusByID(c *gin.Context) {
	id := c.Param("id")
	if !e.isUserOrHasAnyRole(c, id, common.RoleRoot, common.RoleAdmin) {
		return
	}
	u, err := e.repo.GetParticipationStatusByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetParticipationStatusByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (e *EventsAPI) GetAllParticipationStatus(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	skip := c.Query("skip")
	limit := c.Query("limit")
	eventID := c.Query("eventid")
	keycloakID := c.Query("kc_id")

	email := c.Query("email")
	gender := c.Query("gender")
	country := c.Query("country")
	firstName := c.Query("fname")
	lastName := c.Query("lname")
	partOption := c.Query("part-option")
	isCSVReq := c.Query("csv")
	if isCSVReq != "" && isCSVReq != "false" && isCSVReq != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid csv value! Accepted value is either true or false"})
		return
	}

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "100"
	}

	_, sErr := strconv.Atoi(skip)
	if sErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER", "success": false})
		return
	}

	_, lErr := strconv.Atoi(limit)
	if lErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	if isCSVReq == "true" {
		// ALL will fetch all the entries in the DB
		limit = "ALL"
	}

	u, err := e.repo.GetAllParticipationStatus(c.Request.Context(), skip, limit, eventID, keycloakID, country, email, gender,
		partOption, firstName, lastName)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetAllParticipationStatus: %w", err))
		return
	}

	if isCSVReq == "true" {
		c.Writer.Header().Add("Content-Disposition", `attachment; filename=`+time.Now().Format("2006-01-02T15:04:05")+".csv")
		gocsv.Marshal(u, c.Writer)
		c.Status(http.StatusOK)
		return
	} else {
		count, _ := e.repo.GetTotalParticipationStatusCount(c.Request.Context(), eventID, keycloakID, country, email, gender,
			partOption, firstName, lastName)
		c.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "totalCount": count, "success": true})
		return
	}
}

func (e *EventsAPI) CreateNewParticipationStatus(c *gin.Context) {

	s := repo.PartStatusWithNotification{}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := validator.New().Struct(s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	if !e.isUserOrHasAnyRole(c, strconv.Itoa(*s.ParticipantID), common.RoleRoot, common.RoleAdmin) {
		return
	}
	ctx := c.Request.Context()
	id, err := e.repo.CreateNewParticipationStatus(ctx, s)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateNewParticipationStatus: %w", err))
		return
	}

	if s.Notification && s.NotificationType == "confirmation" {
		ps, err := e.repo.GetParticipationStatusByID(ctx, strconv.Itoa(id))
		if err != nil {
			utils.LogFor(ctx).Error("CreateNewParticipationStatus notification [GetParticipationStatusByID]", slog.Any("err", err))
			hub := utils.SentryFor(ctx)
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("participation_status_id", id)
				hub.CaptureMessage("CreateNewParticipationStatus notification [GetParticipationStatusByID]")
			})
		}

		err = utils.SendConfirmationEmail(ctx, *ps.FirstName, *ps.LastName, *ps.Email, *ps.EmailLanguage)
		if err != nil {
			utils.LogFor(ctx).Error("CreateNewParticipationStatus notification [SendConfirmationEmail]", slog.Any("err", err))
			hub := utils.SentryFor(ctx)
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("participation_status_id", id)
				hub.CaptureMessage("CreateNewParticipationStatus notification [SendConfirmationEmail]")
			})
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created new Participation Status!", "data": s, "success": true})
}

func (e *EventsAPI) UpdateParticipationStatus(c *gin.Context) {

	u := repo.PartStatusWithNotification{}
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	id := c.Param("id")
	kcID := c.Param("kcid")
	eventSlug := c.Param("slug")

	if id == "" && (kcID == "" || eventSlug == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters", "success": false})
		return
	}
	if !e.isSubjectOrHasAnyRole(c, kcID, common.RoleRoot, common.RoleAdmin) {
		return
	}
	var err error
	ctx := c.Request.Context()
	if id != "" {
		err = e.repo.UpdateParticipationStatusByID(ctx, u.ParticipationStatusStruct, id)
	} else {
		err = e.repo.UpdateParticipationStatusByKcIDAndEventSlug(ctx, u.ParticipationStatusStruct, kcID, eventSlug)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else if errors.Is(err, common.ErrInvalidValues) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.updateParticipationStatus [byID=%t]: %w", id != "", err))
		}
		return
	}

	if u.Notification && u.NotificationType == "confirmation" {
		ps, err := e.repo.GetParticipationStatusByID(ctx, id)
		if err != nil {
			utils.LogFor(ctx).Error("UpdateParticipationStatus notification [GetParticipationStatusByID]", slog.Any("err", err))
			hub := utils.SentryFor(ctx)
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("participation_status_id", id)
				hub.CaptureMessage("UpdateParticipationStatus notification [GetParticipationStatusByID]")
			})
		}

		err = utils.SendConfirmationEmail(ctx, *ps.FirstName, *ps.LastName, *ps.Email, *ps.EmailLanguage)
		if err != nil {
			utils.LogFor(ctx).Error("UpdateParticipationStatus notification [SendConfirmationEmail]", slog.Any("err", err))
			hub := utils.SentryFor(ctx)
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("participation_status_id", id)
				hub.CaptureMessage("UpdateParticipationStatus notification [SendConfirmationEmail]")
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participation Status updated successfully", "data": u, "success": true})
}

func (e *EventsAPI) DeleteParticipationStatusByID(c *gin.Context) {
	if !e.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	id := c.Param("id")

	if err := e.repo.DeleteParticipationStatusByID(c.Request.Context(), id); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeleteParticipationStatusByID: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participation Status deleted successfully!", "success": true})
}
