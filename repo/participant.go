package repo

import (
	"context"
	"fmt"
	"strings"
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

	if eventId != 0 || eventSlug != "" {
		var eventlOrSlugWhereQuery string
		var eventlOrSlugFromQuery string

		if eventId != 0 {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps`
			eventlOrSlugWhereQuery = fmt.Sprintf(`WHERE ps.event_id = %d 
			AND p.id = ps.participant_id`, eventId)
		} else {
			eventlOrSlugFromQuery = `FROM participant as p, participation_status as ps, event as e`
			eventlOrSlugWhereQuery = fmt.Sprintf(`WHERE e.slug = '%s'
			AND ps.event_id = e.id
			AND p.id = ps.participant_id`, eventSlug)
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

	rows, err := e.Query(ctx, query)
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
	toUpdate, toUpdateArgs := prepareParticipantUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE participant SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewParticipant(ctx context.Context, req Part) (int, error) {
	createString, numString, createQueryArgs := prepareParticipantCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return 0, common.ErrInvalidValues
	}

	var ID int
	if err := e.QueryRow(ctx, fmt.Sprintf(`INSERT INTO participant (%s) VALUES (%s) RETURNING id`, createString, numString),
		createQueryArgs...).Scan(&ID); err != nil {
		return 0, err
	}

	return ID, nil
}

func (e *EventsDB) DeleteParticipantByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from participant where id=$1", id)
	return err
}

func prepareParticipantUpdateQuery(req Part) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.KeycloakID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("keycloak_id=$%d", len(updateStrings)+1))
		args = append(args, *req.KeycloakID)
	}
	if req.FirstLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_language=$%d", len(updateStrings)+1))
		args = append(args, *req.FirstLanguage)
	}
	if req.EmailLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email_language=$%d", len(updateStrings)+1))
		args = append(args, *req.EmailLanguage)
	}
	if req.DOB != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("dob=$%d", len(updateStrings)+1))
		args = append(args, *req.DOB)
	}
	if req.Gender != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("gender=$%d", len(updateStrings)+1))
		args = append(args, *req.Gender)
	}
	if req.Email != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email=$%d", len(updateStrings)+1))
		args = append(args, *req.Email)
	}
	if req.Country != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("country=$%d", len(updateStrings)+1))
		args = append(args, *req.Country)
	}
	if req.PhoneNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("phone_number=$%d", len(updateStrings)+1))
		args = append(args, *req.PhoneNumber)
	}
	if req.FirstName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name=$%d", len(updateStrings)+1))
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name=$%d", len(updateStrings)+1))
		args = append(args, *req.LastName)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareParticipantCreateQuery(req Part) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.KeycloakID != nil {
		createStrings = append(createStrings, "keycloak_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.KeycloakID)
	}
	if req.FirstLanguage != nil {
		createStrings = append(createStrings, "first_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.FirstLanguage)
	}
	if req.EmailLanguage != nil {
		createStrings = append(createStrings, "email_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EmailLanguage)
	}
	if req.DOB != nil {
		createStrings = append(createStrings, "dob")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.DOB)
	}
	if req.Gender != nil {
		createStrings = append(createStrings, "gender")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Gender)
	}
	if req.Email != nil {
		createStrings = append(createStrings, "email")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Email)
	}
	if req.Country != nil {
		createStrings = append(createStrings, "country")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Country)
	}
	if req.PhoneNumber != nil {
		createStrings = append(createStrings, "phone_number")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.PhoneNumber)
	}
	if req.FirstName != nil {
		createStrings = append(createStrings, "first_name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		createStrings = append(createStrings, "last_name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.LastName)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
