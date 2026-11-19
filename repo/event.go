package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type EventResponse struct {
	ID                            *int                                        `json:"id" db:"id"`
	RegistrationRequired          *bool                                       `json:"registration_required" db:"registration_required"`
	RegistrationStatus            *string                                     `json:"registration_status" db:"registration_status"`
	Audience                      *string                                     `json:"audience" db:"audience"`
	Slug                          *string                                     `json:"slug" db:"slug"`
	Name                          *string                                     `json:"name" db:"name"`
	Logo                          *string                                     `json:"logo,omitempty" db:"logo"`
	Content                       *map[string]interface{}                     `json:"content,omitempty" db:"content"`
	Deleted                       *bool                                       `json:"deleted" db:"deleted"`
	StartsOn                      *time.Time                                  `json:"starts_on" db:"starts_on"`
	EndsOn                        *time.Time                                  `json:"ends_on" db:"ends_on"`
	DateConfirmed                 *bool                                       `json:"date_confirmed" db:"date_confirmed"`
	ArchiveLink                   *string                                     `json:"archive_link" db:"archive_link"`
	Published                     *bool                                       `json:"published" db:"published"`
	CreatedAt                     *time.Time                                  `json:"created_at" db:"created_at"`
	UpdatedAt                     *time.Time                                  `json:"updated_at" db:"updated_at"`
	IsUserRegistered              *bool                                       `json:"is_user_registered,omitempty"`
	ParticipationOption           []EventPartOptionResponse                   `json:"participation_options,omitempty"`
	UserParticipationOptionDetail ParticipationStatusStructWithCreationDetail `json:"user_participation_details,omitempty"`
}

type Event struct {
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status"`
	Audience             *string    `json:"audience" db:"audience"`
	Slug                 *string    `json:"slug" db:"slug" validate:"required"`
	Name                 *string    `json:"name" db:"name" validate:"required"`
	Logo                 *string    `json:"logo,omitempty" db:"logo"`
	Content              *string    `json:"content,omitempty" db:"content"`
	Deleted              *bool      `json:"deleted" db:"deleted"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on" validate:"required"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on" validate:"required"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed"`
	ArchiveLink          *string    `json:"archive_link" db:"archive_link"`
	Published            *bool      `json:"published" db:"published"`
}

func (e *EventsDB) GetEventByID(ctx context.Context, id string) (*EventResponse, error) {
	u := EventResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	deleted,
	starts_on,
	ends_on,
	date_confirmed,
	archive_link,
	published,
	created_at,
	updated_at 
	from event where id = $1`, id).Scan(
		&u.ID,
		&u.RegistrationRequired,
		&u.RegistrationStatus,
		&u.Audience,
		&u.Slug,
		&u.Name,
		&u.Logo,
		&u.Content,
		&u.Deleted,
		&u.StartsOn,
		&u.EndsOn,
		&u.DateConfirmed,
		&u.ArchiveLink,
		&u.Published,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("e.QueryRow: %w", err)
	}

	// Attach participation options for the event.
	rows, err := e.Query(ctx, `select participation_option from event_participation_option where event_id = $1`, u.ID)
	if err != nil {
		return nil, fmt.Errorf("e.Query [participation_option]: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var epoResponse EventPartOptionResponse
		err = rows.Scan(&epoResponse.ParticipationOption)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		u.ParticipationOption = append(u.ParticipationOption, epoResponse)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return &u, nil
}

func (e *EventsDB) GetAllEvent(ctx context.Context, skip int, limit int, slug string, email string, kcID string) ([]EventResponse, error) {
	var query string
	var args sqlArgs

	if email != "" || kcID != "" {
		// Both subqueries below filter on the same $1.
		var emailOrKcQuery string
		if email != "" {
			emailOrKcQuery = "where p.email=$1"
			args = append(args, email)
		} else {
			emailOrKcQuery = "where p.keycloak_id=$1"
			args = append(args, kcID)
		}

		query = fmt.Sprintf(`select 
			e.id,
			e.registration_required,
			e.registration_status,
			e.audience,
			e.slug,
			e.name,
			e.logo,
			e.content,
			e.deleted,
			e.starts_on,
			e.ends_on,
			e.date_confirmed,
			e.archive_link,
			e.published,
			e.created_at,
			e.updated_at,
			participation_status.participation_option,
			participation_status.confirmed,
			participation_status.registration_date,
			participation_status.deleted,
			participation_status.created_at,
			participation_status.updated_at,
			CASE WHEN (SELECT COUNT(*) FROM participation_status as ps, participant as p %s AND ps.participant_id = p.id AND e.id = ps.event_id ) > 0 THEN true
			ELSE false
			END AS is_user_registered 
			from event as e
			LEFT JOIN participation_status ON participation_status.participant_id = ( SELECT id FROM participant as p %s) AND participation_status.event_id = e.id
			WHERE e.deleted = false
			LIMIT %d OFFSET %d`, emailOrKcQuery, emailOrKcQuery, limit, skip)
	} else {
		whereQuery := buildAndGetWhereEventQuery(&args, slug)

		query = fmt.Sprintf(`select 
		id,
		registration_required,
		registration_status,
		audience,
		slug,
		name,
		logo,
		content,
		deleted,
		starts_on,
		ends_on,
		date_confirmed,
		archive_link,
		published,
		created_at,
		updated_at 
		from event`+whereQuery+" LIMIT %d OFFSET %d", limit, skip)
	}

	rows, err := e.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []EventResponse{}
	for rows.Next() {
		var d EventResponse
		var err error

		// Applied these checks to handle extra output is_user_registered when email or kcID is passed
		if email != "" || kcID != "" {
			err = rows.Scan(&d.ID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name,
				&d.Logo, &d.Content, &d.Deleted, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.ArchiveLink, &d.Published,
				&d.CreatedAt, &d.UpdatedAt, &d.UserParticipationOptionDetail.ParticipationOption, &d.UserParticipationOptionDetail.Confirmed,
				&d.UserParticipationOptionDetail.RegistrationDate, &d.UserParticipationOptionDetail.Deleted,
				&d.UserParticipationOptionDetail.CreatedAt, &d.UserParticipationOptionDetail.UpdatedAt, &d.IsUserRegistered)
		} else {
			err = rows.Scan(&d.ID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo,
				&d.Content, &d.Deleted, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.ArchiveLink, &d.Published, &d.CreatedAt, &d.UpdatedAt)
		}
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		// Attach participation options for the event.
		epo, err := e.Query(ctx, `select participation_option from event_participation_option where event_id = $1`, d.ID)
		if err != nil {
			return nil, fmt.Errorf("e.Query [participation_option]: %w", err)
		}
		defer epo.Close()

		for epo.Next() {
			var epoResponse EventPartOptionResponse
			err = epo.Scan(&epoResponse.ParticipationOption)
			if err != nil {
				return nil, fmt.Errorf("epo.Scan: %w", err)
			}
			d.ParticipationOption = append(d.ParticipationOption, epoResponse)
		}
		if epo.Err(); err != nil {
			return nil, fmt.Errorf("epo.Err: %w", err)
		}

		u = append(u, d)
	}
	if rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return u, nil
}

func (e *EventsDB) UpdateEventByID(ctx context.Context, req Event, id string) error {
	c := eventUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("event", "id", id)
	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateEvent(ctx context.Context, req Event) error {
	c := eventCreateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("event")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventByID(ctx context.Context, id string) error {
	// The four updates used to go as one multi-statement string, which
	// Postgres runs as a single transaction; keep that with an explicit tx.
	tx, err := e.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, q := range []string{
		`UPDATE event SET deleted = true WHERE id=$1`,
		`UPDATE event_item SET deleted = true WHERE event_id=$1`,
		`UPDATE event_participation_option SET deleted = true WHERE event_id=$1`,
		`UPDATE participation_status SET deleted = true WHERE event_id=$1`,
	} {
		if _, err := tx.Exec(ctx, q, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (e *EventsDB) DeleteHardEventByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event where id=$1", id)
	return err
}

func eventUpdateColumns(req Event) columns {
	var c columns
	setIf(&c, "registration_required", req.RegistrationRequired)
	setIf(&c, "registration_status", req.RegistrationStatus)
	setIf(&c, "audience", req.Audience)
	setIf(&c, "slug", req.Slug)
	setIf(&c, "name", req.Name)
	setIf(&c, "logo", req.Logo)
	setIf(&c, "content", req.Content)
	setIf(&c, "deleted", req.Deleted)
	setIf(&c, "starts_on", req.StartsOn)
	setIf(&c, "ends_on", req.EndsOn)
	setIf(&c, "date_confirmed", req.DateConfirmed)
	setIf(&c, "archive_link", req.ArchiveLink)
	setIf(&c, "published", req.Published)
	return c
}

func eventCreateColumns(req Event) columns {
	var c columns
	setIf(&c, "registration_required", req.RegistrationRequired)
	setIf(&c, "registration_status", req.RegistrationStatus)
	setIf(&c, "audience", req.Audience)
	setIf(&c, "slug", req.Slug)
	setIf(&c, "name", req.Name)
	if req.Logo != nil && *req.Logo != "" {
		c.set("logo", *req.Logo)
	}
	if req.Content != nil && *req.Content != "" {
		c.set("content", *req.Content)
	}
	setIf(&c, "deleted", req.Deleted)
	setIf(&c, "starts_on", req.StartsOn)
	setIf(&c, "ends_on", req.EndsOn)
	setIf(&c, "date_confirmed", req.DateConfirmed)
	setIf(&c, "archive_link", req.ArchiveLink)
	setIf(&c, "published", req.Published)
	return c
}

func buildAndGetWhereEventQuery(args *sqlArgs, slug string) string {

	var whereString strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	whereCondition.WriteString(" deleted=false")

	// WHERE query generation based on parameters
	if slug != "" {
		whereCondition.WriteString(" AND slug=" + args.add(slug))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String()
}
