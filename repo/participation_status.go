package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type ParticipationStatusResponse struct {
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

type PartStatusNotification struct {
	Notification     bool   `json:"notification"`
	NotificationType string `json:"notification_type"`
}

type PartStatusWithNotification struct {
	ParticipationStatusStruct
	PartStatusNotification
}

func (e *EventsDB) GetParticipationStatusByID(ctx context.Context, id string) (*ParticipationStatusResponse, error) {
	u := ParticipationStatusResponse{}
	if err := e.QueryRow(ctx, `select 
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
		return nil, err
	}
	return &u, nil
}

func (e *EventsDB) GetAllParticipationStatus(ctx context.Context, skip string, limit string, eventID string, keycloakID string, country string,
	email string, gender string, partOption string, firstName string, lastName string) ([]ParticipationStatusResponse, error) {
	var args sqlArgs
	userDbWhereQuery, orderByQuery := buildAndGetWhereQuery(&args, eventID, keycloakID, country, email, gender, partOption, firstName, lastName)
	// "ALL" (CSV export) means no limit; it is not a number to bind.
	limitOffsetString := " OFFSET " + args.add(skip)
	if limit != "ALL" {
		limitOffsetString = " LIMIT " + args.add(limit) + limitOffsetString
	}

	rows, err := e.Query(ctx, `select 
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
		orderByQuery+limitOffsetString, args...)
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []ParticipationStatusResponse{}
	for rows.Next() {
		var d ParticipationStatusResponse
		err := rows.Scan(
			&d.ID, &d.ParticipationOption, &d.ParticipantID, &d.EventID, &d.Confirmed, &d.RegistrationDate, &d.Deleted, &d.CreatedAt, &d.UpdatedAt,
			&d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.StartsOn, &d.EndsOn, &d.DateConfirmed,
			&d.KeycloakID, &d.FirstLanguage, &d.EmailLanguage, &d.DOB, &d.Gender, &d.Email, &d.PhoneNumber, &d.Country, &d.FirstName, &d.LastName,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		u = append(u, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return u, nil
}

func (e *EventsDB) UpdateParticipationStatusByID(ctx context.Context, req ParticipationStatusStruct, id string) error {
	c := participationStatusUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("participation_status", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("o.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) UpdateParticipationStatusByKcIDAndEventSlug(ctx context.Context, req ParticipationStatusStruct, kcID string, eventSlug string) error {
	c := participationStatusUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())

	query := `UPDATE participation_status SET ` + c.setClause() + ` WHERE participant_id=(SELECT id FROM participant as p 
		WHERE p.keycloak_id=` + c.args.add(kcID) + `) AND event_id=(SELECT id FROM event as e WHERE e.slug=` + c.args.add(eventSlug) + `)`
	updateRes, err := e.Exec(ctx, query, c.args...)
	if err != nil {
		return fmt.Errorf("o.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewParticipationStatus(ctx context.Context, req PartStatusWithNotification) (int, error) {
	c := participationStatusCreateColumns(req.ParticipationStatusStruct)
	if c.empty() {
		return 0, common.ErrInvalidValues
	}
	query, args := c.insert("participation_status")

	var id int

	if err := e.QueryRow(ctx, query+` RETURNING id`, args...).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("no rows affected")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// If the error is about a constraint violation,
			if pgErr.Code == "23505" {
				if err := e.QueryRow(ctx, `UPDATE participation_status SET participation_option=$1,updated_at=$2 WHERE participant_id=$3 AND event_id=$4 RETURNING id`,
					*req.ParticipationStatusStruct.ParticipationOption, time.Now(), *req.ParticipationStatusStruct.ParticipantID, req.ParticipationStatusStruct.EventID).
					Scan(&id); err != nil {
					return 0, fmt.Errorf("e.QueryRow.Scan [pgErr.Code 23505]: %w", err)
				}
			} else {
				return 0, fmt.Errorf("pgconn error: %w", err)
			}
		} else {
			return 0, fmt.Errorf("e.QueryRow.Scan: %w", err)
		}
	}

	return id, nil
}

func (e *EventsDB) GetTotalParticipationStatusCount(ctx context.Context, eventID string, keycloakID string, country string,
	email string, gender string, partOption string, firstName string, lastName string) (int, error) {
	var count int

	var args sqlArgs
	userDbWhereQuery, _ := buildAndGetWhereQuery(&args, eventID, keycloakID, country, email, gender, partOption, firstName, lastName)

	err := e.QueryRow(ctx, `SELECT COUNT(*) FROM participation_status 
	LEFT JOIN participant ON participation_status.participant_id = participant.id`+userDbWhereQuery, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (e *EventsDB) DeleteParticipationStatusByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from participation_status where id=$1", id)
	return err
}

func participationStatusUpdateColumns(req ParticipationStatusStruct) columns {
	var c columns
	setIf(&c, "participation_option", req.ParticipationOption)
	setIf(&c, "participant_id", req.ParticipantID)
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "confirmed", req.Confirmed)
	setIf(&c, "registration_date", req.RegistrationDate)
	setIf(&c, "deleted", req.Deleted)
	return c
}

func participationStatusCreateColumns(req ParticipationStatusStruct) columns {
	var c columns
	setIf(&c, "participation_option", req.ParticipationOption)
	setIf(&c, "participant_id", req.ParticipantID)
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "confirmed", req.Confirmed)
	setIf(&c, "registration_date", req.RegistrationDate)
	setIf(&c, "deleted", req.Deleted)
	return c
}

func buildAndGetWhereQuery(args *sqlArgs, eventID string, keycloakID string, country string, email string, gender string, partOption string, firstName string, lastName string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	//deleted false query
	whereCondition.WriteString(" participation_status.deleted=false")

	// WHERE query generation based on parameters
	if eventID != "" {
		whereCondition.WriteString(" AND event_id=" + args.add(eventID))
	}

	if country != "" {
		whereCondition.WriteString(" AND LOWER(participant.country)=LOWER(" + args.add(country) + ")")
	}

	if email != "" {
		whereCondition.WriteString(" AND LOWER(participant.email) LIKE LOWER('%' || " + args.add(email) + " || '%')")
	}

	if gender != "" {
		whereCondition.WriteString(" AND LOWER(participant.gender)=LOWER(" + args.add(gender) + ")")
	}

	if partOption != "" {
		whereCondition.WriteString(" AND LOWER(participation_option) LIKE LOWER('%' || " + args.add(partOption) + " || '%')")
	}

	if firstName != "" {
		whereCondition.WriteString(" AND LOWER(participant.first_name) LIKE LOWER('%' || " + args.add(firstName) + " || '%')")
	}

	if lastName != "" {
		whereCondition.WriteString(" AND LOWER(participant.last_name) LIKE LOWER('%' || " + args.add(lastName) + " || '%')")
	}

	if keycloakID != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(" AND participant.keycloak_id=" + args.add(keycloakID))
		} else {
			whereCondition.WriteString(" participant.keycloak_id=" + args.add(keycloakID))
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
