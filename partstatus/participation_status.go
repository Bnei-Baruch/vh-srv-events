package partstatus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vh-srv-event/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type participationStatusResponse struct {
	ID                  *int       `json:"id" db:"id"`
	ParticipationOption *string    `json:"participation_option" db:"participation_option"`
	ParticipantID       *int       `json:"participant_id" db:"participant_id"`
	EventID             *int       `json:"event_id" db:"event_id"`
	Confirmed           *bool      `json:"confirmed" db:"confirmed"`
	RegistrationDate    *time.Time `json:"registration_date" db:"registration_date"`
	Deleted             *bool      `json:"deleted" db:"deleted"`
	CreatedAt           *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at" db:"updated_at"`
	//Event
	RegistrationRequired *bool      `json:"event_registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"event_registration_status" db:"registration_status"`
	Audience             *string    `json:"event_audience" db:"audience"`
	Slug                 *string    `json:"event_slug" db:"slug"`
	Name                 *string    `json:"event_name" db:"name"`
	Logo                 *string    `json:"event_logo,omitempty" db:"logo"`
	Content              *string    `json:"event_content,omitempty" db:"content"`
	StartsOn             *time.Time `json:"event_starts_on" db:"starts_on"`
	EndsOn               *time.Time `json:"event_ends_on" db:"ends_on"`
	DateConfirmed        *bool      `json:"event_date_confirmed" db:"date_confirmed"`
	//Participant
	KeycloakID    *string    `json:"part_keycloak_id" db:"keycloak_id"`
	FirstLanguage *string    `json:"part_first_language,omitempty" db:"first_language"`
	EmailLanguage *string    `json:"part_email_language,omitempty" db:"email_language"`
	DOB           *time.Time `json:"part_dob,omitempty" db:"dob"`
	Gender        *string    `json:"part_gender,omitempty" db:"gender"`
	Email         *string    `json:"part_email" db:"email"`
	PhoneNumber   *string    `json:"part_phone_number" db:"phone_number"`
	Country       *string    `json:"part_country,omitempty" db:"country"`
	FirstName     *string    `json:"part_first_name" db:"first_name"`
	LastName      *string    `json:"part_last_name" db:"last_name"`
}

type ParticipationStatusStruct struct {
	ParticipationOption *string    `json:"participation_option,omitempty" db:"participation_option" validate:"required"`
	ParticipantID       *int       `json:"participant_id,omitempty" db:"participant_id" validate:"required"`
	EventID             *int       `json:"event_id,omitempty" db:"event_id" validate:"required"`
	Confirmed           *bool      `json:"confirmed,omitempty" db:"confirmed"`
	RegistrationDate    *time.Time `json:"registration_date,omitempty" db:"registration_date" validate:"required"`
	Deleted             *bool      `json:"deleted,omitempty" db:"deleted"`
}

type ParticipationStatusStructWithCreationDetail struct {
	ParticipationStatusStruct
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type partStatusNotification struct {
	Notification     bool   `json:"notification"`
	NotificationType string `json:"notification_type"`
}

type partStatusWithNotification struct {
	ParticipationStatusStruct
	partStatusNotification
}

type ParticipationStatus interface {
	GetParticipationStatusByID(ctx *gin.Context)
	GetAllParticipationStatus(ctx *gin.Context)
	CreateNewParticipationStatus(ctx *gin.Context)
	UpdateParticipationStatus(ctx *gin.Context)
	DeleteParticipationStatusByID(ctx *gin.Context)
}

type ParticipationStatusDB struct {
	db *pgxpool.Pool
}

func NewParticipationStatus(db *pgxpool.Pool) ParticipationStatus {
	return &ParticipationStatusDB{
		db,
	}
}

func (r *ParticipationStatusDB) GetParticipationStatusByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getParticipationStatusByID(r, ctx, id)

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

func (r *ParticipationStatusDB) GetAllParticipationStatus(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")
	eventID := ctx.Query("eventid")
	keycloakID := ctx.Query("kc_id")

	email := ctx.Query("email")
	gender := ctx.Query("gender")
	country := ctx.Query("country")
	firstName := ctx.Query("fname")
	lastName := ctx.Query("lname")
	partOption := ctx.Query("part-option")
	isCSVReq := ctx.Query("csv")
	if isCSVReq != "" && isCSVReq != "false" && isCSVReq != "true" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid csv value! Accepted value is either true or false"})
		return
	}

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "10"
	}

	// String conversion to int
	_, sErr := strconv.Atoi(skip)
	if sErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER", "success": false})
		return
	}

	// String conversion to int
	_, lErr := strconv.Atoi(limit)
	if lErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	if isCSVReq == "true" {
		// ALL will fetch all the entries in the DB
		limit = "ALL"
	}

	u, err := getAllParticipationStatus(r, ctx, skip, limit, eventID, keycloakID, country, email, gender, partOption, firstName, lastName)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if isCSVReq == "true" {
		ctx.Writer.Header().Add("Content-Disposition", `attachment; filename=`+time.Now().Format("2006-01-02T15:04:05")+".csv")
		gocsv.Marshal(u, ctx.Writer)
		ctx.Status(http.StatusOK)
		return
	} else {
		count, _ := getTotalParticipationStatusCount(r, ctx, eventID, keycloakID, country, email, gender, partOption, firstName, lastName)
		ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "totalCount": count, "success": true})
		return
	}
}

func (r *ParticipationStatusDB) CreateNewParticipationStatus(ctx *gin.Context) {
	s := partStatusWithNotification{}
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

	id, err := createNewParticipationStatus(r, ctx, s)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if s.Notification && s.NotificationType == "confirmation" {
		emailErr := util.SendConfirmationEmail(ctx, r.db, id)
		if emailErr != nil {
			fmt.Println("registered user to event but problem sending email: %w", emailErr)
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Participation Status!", "data": s, "success": true})
}

func (r *ParticipationStatusDB) UpdateParticipationStatus(ctx *gin.Context) {
	u := partStatusWithNotification{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")
	kcID := ctx.Param("kcid")
	eventSlug := ctx.Param("slug")

	if id == "" && (kcID == "" || eventSlug == "") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameters",
			"success": false,
		})
		return
	}

	if id != "" {
		if err := updateParticipationStatusByID(r, ctx, u.ParticipationStatusStruct, id); err != nil {

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
	} else {

		if err := updateParticipationStatusByKcIDAndEventSlug(r, ctx, u.ParticipationStatusStruct, kcID, eventSlug); err != nil {

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
	}

	if u.Notification && u.NotificationType == "confirmation" {
		intID, _ := strconv.Atoi(id)
		emailErr := util.SendConfirmationEmail(ctx, r.db, intID)
		if emailErr != nil {
			fmt.Println("problem sending email in update participation status: %w", emailErr)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participation Status updated successfully", "data": u, "success": true})
}

func (r *ParticipationStatusDB) DeleteParticipationStatusByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteParticipationStatusByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participation Status deleted successfully!", "success": true})
}

func getParticipationStatusByID(r *ParticipationStatusDB, ctx *gin.Context, id string) (participationStatusResponse, error) {
	u := participationStatusResponse{}
	if err := r.db.QueryRow(ctx, `select 
	participation_status.id,
	participation_option,
	participant_id,
	event_id,
	confirmed,
	registration_date,
	participation_status.deleted,
	participation_status.created_at,
	participation_status.updated_at,
	event.registration_required,
	event.registration_status,
	event.audience,
	event.slug,
	event.name,
	event.logo,
	event.content,
	event.starts_on,
	event.ends_on,
	event.date_confirmed,
	participant.keycloak_id,
	participant.first_language,
	participant.email_language,
	participant.dob,
	participant.gender,
	participant.email,
	participant.phone_number,
	participant.country,
	participant.first_name,
	participant.last_name
	FROM participation_status 
	LEFT JOIN event ON participation_status.event_id = event.id LEFT JOIN participant ON participation_status.participant_id = participant.id where participation_status.id = $1`, id).Scan(
		&u.ID, &u.ParticipationOption, &u.ParticipantID, &u.EventID, &u.Confirmed, &u.RegistrationDate, &u.Deleted, &u.CreatedAt, &u.UpdatedAt,
		&u.RegistrationRequired, &u.RegistrationStatus, &u.Audience, &u.Slug, &u.Name, &u.Logo, &u.Content, &u.StartsOn, &u.EndsOn, &u.DateConfirmed,
		&u.KeycloakID, &u.FirstLanguage, &u.EmailLanguage, &u.DOB, &u.Gender, &u.Email, &u.PhoneNumber, &u.Country, &u.FirstName, &u.LastName,
	); err != nil {
		if err == pgx.ErrNoRows {
			return participationStatusResponse{}, fmt.Errorf("not found")
		}
		return participationStatusResponse{}, err
	}
	return u, nil
}

func getAllParticipationStatus(r *ParticipationStatusDB, ctx *gin.Context, skip string, limit string, eventID string, keycloakID string, country string, email string, gender string, partOption string, firstName string, lastName string) (*[]participationStatusResponse, error) {

	u := []participationStatusResponse{}

	limitOffsetString := fmt.Sprintf(" LIMIT %s OFFSET %s", limit, skip)

	userDbWhereQuery, orderByQuery := buildAndGetWhereQuery(eventID, keycloakID, country, email, gender, partOption, firstName, lastName)

	rows, err := r.db.Query(ctx, `select 
	participation_status.id,
	participation_option,
	participant_id,
	event_id,
	confirmed,
	registration_date,
	participation_status.deleted,
	participation_status.created_at,
	participation_status.updated_at,
	event.registration_required,
	event.registration_status,
	event.audience,
	event.slug,
	event.name,
	event.logo,
	event.content,
	event.starts_on,
	event.ends_on,
	event.date_confirmed,
	participant.keycloak_id,
	participant.first_language,
	participant.email_language,
	participant.dob,
	participant.gender,
	participant.email,
	participant.phone_number,
	participant.country,
	participant.first_name,
	participant.last_name
	FROM participation_status 
	LEFT JOIN event ON participation_status.event_id = event.id LEFT JOIN participant ON participation_status.participant_id = participant.id`+userDbWhereQuery+
		orderByQuery+limitOffsetString)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return &u, err
	}
	defer rows.Close()
	for rows.Next() {
		var d participationStatusResponse
		err := rows.Scan(
			&d.ID, &d.ParticipationOption, &d.ParticipantID, &d.EventID, &d.Confirmed, &d.RegistrationDate, &d.Deleted, &d.CreatedAt, &d.UpdatedAt,
			&d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.StartsOn, &d.EndsOn, &d.DateConfirmed,
			&d.KeycloakID, &d.FirstLanguage, &d.EmailLanguage, &d.DOB, &d.Gender, &d.Email, &d.PhoneNumber, &d.Country, &d.FirstName, &d.LastName,
		)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateParticipationStatusByID(r *ParticipationStatusDB, ctx *gin.Context, req ParticipationStatusStruct, id string) error {

	toUpdate, toUpdateArgs := prepareParticipationStatusUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE participation_status SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating Participation Status: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func updateParticipationStatusByKcIDAndEventSlug(r *ParticipationStatusDB, ctx *gin.Context, req ParticipationStatusStruct, kcID string, eventSlug string) error {

	toUpdate, toUpdateArgs := prepareParticipationStatusUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE participation_status SET %s WHERE participant_id=(SELECT id FROM participant as p WHERE p.keycloak_id='%s') AND event_id=(SELECT id FROM event as e WHERE e.slug='%s')`, toUpdate, kcID, eventSlug),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating Participation Status: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewParticipationStatus(r *ParticipationStatusDB, ctx *gin.Context, req partStatusWithNotification) (int, error) {

	createString, numString, createQueryArgs := prepareParticipationStatusCreateQuery(req.ParticipationStatusStruct)

	var id int
	if len(createQueryArgs) != 0 {
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO participation_status (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(
			&id,
		); err != nil {
			if err == pgx.ErrNoRows {
				return id, fmt.Errorf("no rows affected")
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Message)
				// If the error is about a constraint violation,
				if pgErr.Code == "23505" {
					if err := r.db.QueryRow(ctx, `UPDATE participation_status SET participation_option=$1,updated_at=$2 WHERE participant_id=$3 AND event_id=$4 RETURNING id`,
						*req.ParticipationStatusStruct.ParticipationOption, time.Now(), *req.ParticipationStatusStruct.ParticipantID, req.ParticipationStatusStruct.EventID).Scan(
						&id,
					); err != nil {
						return id, fmt.Errorf("problem updating Participation Status: %w", err)
					}
				} else {
					return id, fmt.Errorf("problem inserting Participation Status: %w", err)
				}
			} else {
				return id, fmt.Errorf("problem inserting Participation Status: %w", err)
			}
		}

		return id, nil
	} else {
		return id, fmt.Errorf("invalid values")
	}
}

func getTotalParticipationStatusCount(r *ParticipationStatusDB, ctx *gin.Context, eventID string, keycloakID string, country string, email string, gender string, partOption string, firstName string, lastName string) (int, error) {
	var count int

	userDbWhereQuery, _ := buildAndGetWhereQuery(eventID, keycloakID, country, email, gender, partOption, firstName, lastName)

	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM participation_status 
	LEFT JOIN participant ON participation_status.participant_id = participant.id`+userDbWhereQuery).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func deleteParticipationStatusByID(r *ParticipationStatusDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from participation_status where id=$1", id)
	return err
}

func prepareParticipationStatusUpdateQuery(req ParticipationStatusStruct) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.ParticipationOption != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participation_option=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.ParticipantID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participant_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipantID)
	}
	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
	}
	if req.Confirmed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("confirmed=$%d", len(updateStrings)+1))
		args = append(args, *req.Confirmed)
	}
	if req.RegistrationDate != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_date=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationDate)
	}
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
		// Updating the deleted_at column in the database if deleted to true. # Implementation pending
		// if *req.Deleted {
		// 	updateStrings = append(updateStrings, fmt.Sprintf("deleted_at=$%d", len(updateStrings)+1))
		// 	args = append(args, time.Now())
		// }
	}
	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareParticipationStatusCreateQuery(req ParticipationStatusStruct) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.ParticipationOption != nil {
		createStrings = append(createStrings, "participation_option")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.ParticipantID != nil {
		createStrings = append(createStrings, "participant_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipantID)
	}
	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.Confirmed != nil {
		createStrings = append(createStrings, "confirmed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Confirmed)
	}
	if req.RegistrationDate != nil {
		createStrings = append(createStrings, "registration_date")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationDate)
	}
	if req.Deleted != nil {
		createStrings = append(createStrings, "deleted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Deleted)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func buildAndGetWhereQuery(eventID string, keycloakID string, country string, email string, gender string, partOption string, firstName string, lastName string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	//deleted false query
	whereCondition.WriteString(" participation_status.deleted=false")

	// WHERE query generation based on parameters
	if eventID != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND event_id=%s", eventID))
	}

	if country != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participant.country)=LOWER('%s')", country))
	}

	if email != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participant.email) LIKE LOWER('%%%s%%')", email))
	}

	if gender != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participant.gender)=LOWER('%s')", gender))
	}

	if partOption != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participation_option) LIKE LOWER('%%%s%%')", partOption))
	}

	if firstName != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participant.first_name) LIKE LOWER('%%%s%%')", firstName))
	}

	if lastName != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND LOWER(participant.last_name) LIKE LOWER('%%%s%%')", lastName))
	}

	if keycloakID != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND participant.keycloak_id='%s'", keycloakID))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" participant.keycloak_id='%s'", keycloakID))
		}
	}

	orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", "desc"))

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
