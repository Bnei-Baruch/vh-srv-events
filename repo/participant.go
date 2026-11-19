package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type PartResponse struct {
	ID            *int       `json:"id" db:"id"`
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	FirstLanguage *string    `json:"first_language,omitempty" db:"first_language"`
	EmailLanguage *string    `json:"email_language,omitempty" db:"email_language"`
	DOB           *time.Time `json:"dob,omitempty" db:"dob"`
	Gender        *string    `json:"gender,omitempty" db:"gender"`
	Email         *string    `json:"email" db:"email"`
	Country       *string    `json:"country,omitempty" db:"country"`
	PhoneNumber   *string    `json:"phone_number,omitempty" db:"phone_number"`
	FirstName     *string    `json:"first_name" db:"first_name"`
	LastName      *string    `json:"last_name" db:"last_name"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

type Part struct {
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id" validate:"required,uuid"`
	FirstLanguage *string    `json:"first_language,omitempty" db:"first_language"`
	EmailLanguage *string    `json:"email_language,omitempty" db:"email_language"`
	DOB           *time.Time `json:"dob,omitempty" db:"dob"`
	Gender        *string    `json:"gender,omitempty" db:"gender"`
	Email         *string    `json:"email" db:"email" validate:"required,email"`
	Country       *string    `json:"country,omitempty" db:"country"`
	PhoneNumber   *string    `json:"phone_number,omitempty" db:"phone_number"`
	FirstName     *string    `json:"first_name" db:"first_name" validate:"required"`
	LastName      *string    `json:"last_name" db:"last_name" validate:"required"`
}

func (e *EventsDB) GetParticipantById(ctx context.Context, id string) (*PartResponse, error) {
	u := PartResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	phone_number,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where id = $1`, id).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.PhoneNumber,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetParticipantByEmail(ctx context.Context, email string) (*PartResponse, error) {
	u := PartResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	phone_number,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where email = $1`, email).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.PhoneNumber,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetParticipantByKeycloakID(ctx context.Context, id string) (*PartResponse, error) {
	u := PartResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	keycloak_id,
	first_language,
	email_language,
	dob,
	gender,
	email,
	country,
	phone_number,
	first_name,
	last_name,
	created_at,
	updated_at 
	from participant where keycloak_id = $1`, id).Scan(
		&u.ID,
		&u.KeycloakID,
		&u.FirstLanguage,
		&u.EmailLanguage,
		&u.DOB,
		&u.Gender,
		&u.Email,
		&u.Country,
		&u.PhoneNumber,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllParticipants(ctx context.Context, skip int, limit int, eventId int, eventSlug string) ([]PartResponse, error) {
	var query string
	var args []interface{}

	if eventId != 0 || eventSlug != "" {
		var eventlOrSlugWhereQuery string
		var eventlOrSlugFromQuery string

		if eventId != 0 {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps`
			eventlOrSlugWhereQuery = fmt.Sprintf(`WHERE ps.event_id = %d 
			AND p.id = ps.participant_id`, eventId)
		} else {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps, event as e`
			eventlOrSlugWhereQuery = `WHERE e.slug = $1
			AND ps.event_id = e.id
			AND p.id = ps.participant_id`
			args = append(args, eventSlug)
		}

		query = fmt.Sprintf(`SELECT 
			p.id,
			p.keycloak_id,
			p.first_language,
			p.email_language,
			p.dob,
			p.gender,
			p.email,
			p.country,
			p.phone_number,
			p.first_name,
			p.last_name,
			p.created_at,
			p.updated_at 
			%s
			%s
			LIMIT %d OFFSET %d`, eventlOrSlugFromQuery, eventlOrSlugWhereQuery, limit, skip)

	} else {
		query = fmt.Sprintf(`SELECT 
			id,
			keycloak_id,
			first_language,
			email_language,
			dob,
			gender,
			email,
			country,
			phone_number,
			first_name,
			last_name,
			created_at,
			updated_at 
			FROM participant LIMIT %d OFFSET %d`, limit, skip)
	}

	rows, err := e.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []PartResponse{}
	for rows.Next() {
		var d PartResponse
		err := rows.Scan(&d.ID, &d.KeycloakID, &d.FirstLanguage, &d.EmailLanguage, &d.DOB, &d.Gender, &d.Email,
			&d.Country, &d.PhoneNumber, &d.FirstName, &d.LastName, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateParticipantByID(ctx context.Context, req Part, id string) error {
	c := participantUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("participant", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewParticipant(ctx context.Context, req Part) (int, error) {
	c := participantCreateColumns(req)
	if c.empty() {
		return 0, common.ErrInvalidValues
	}
	query, args := c.insert("participant")

	var ID int
	if err := e.QueryRow(ctx, query+` RETURNING id`, args...).Scan(&ID); err != nil {
		return 0, err
	}

	return ID, nil
}

func (e *EventsDB) DeleteParticipantByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from participant where id=$1", id)
	return err
}

func (e *EventsDB) IsSubjectID(ctx context.Context, keycloakID, accountID string) (bool, error) {
	row := e.QueryRow(ctx, "SELECT 1 FROM participant WHERE keycloak_id = $1 AND id = $2", keycloakID, accountID)
	var x int
	if err := row.Scan(&x); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func participantUpdateColumns(req Part) columns {
	var c columns
	setIf(&c, "keycloak_id", req.KeycloakID)
	setIf(&c, "first_language", req.FirstLanguage)
	setIf(&c, "email_language", req.EmailLanguage)
	setIf(&c, "dob", req.DOB)
	setIf(&c, "gender", req.Gender)
	setIf(&c, "email", req.Email)
	setIf(&c, "country", req.Country)
	setIf(&c, "phone_number", req.PhoneNumber)
	setIf(&c, "first_name", req.FirstName)
	setIf(&c, "last_name", req.LastName)
	return c
}

func participantCreateColumns(req Part) columns {
	var c columns
	setIf(&c, "keycloak_id", req.KeycloakID)
	setIf(&c, "first_language", req.FirstLanguage)
	setIf(&c, "email_language", req.EmailLanguage)
	setIf(&c, "dob", req.DOB)
	setIf(&c, "gender", req.Gender)
	setIf(&c, "email", req.Email)
	setIf(&c, "country", req.Country)
	setIf(&c, "phone_number", req.PhoneNumber)
	setIf(&c, "first_name", req.FirstName)
	setIf(&c, "last_name", req.LastName)
	return c
}
