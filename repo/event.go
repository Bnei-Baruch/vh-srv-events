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

	if email != "" || kcID != "" {
		var emailOrKcQuery string
		if email != "" {
			emailOrKcQuery = fmt.Sprintf("where p.email='%s'", email)
		} else {
			emailOrKcQuery = fmt.Sprintf("where p.keycloak_id='%s'", kcID)
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
		whereQuery := buildAndGetWhereEventQuery(slug)

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

	rows, err := e.Query(ctx, query)
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
	toUpdate, toUpdateArgs := prepareEventUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}
	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE event SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateEvent(ctx context.Context, req Event) error {
	createString, numString, createQueryArgs := prepareEventCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO event (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventByID(ctx context.Context, id string) error {
	eventQuery := "UPDATE event SET deleted = true WHERE id=" + id + ";"
	eventItemQuery := "UPDATE event_item SET deleted = true WHERE event_id=" + id + ";"
	eventPartQuery := "UPDATE event_participation_option SET deleted = true WHERE event_id=" + id + ";"
	eventStatusQuery := "UPDATE participation_status SET deleted = true WHERE event_id=" + id + ";"
	_, err := e.Exec(ctx, eventQuery+eventItemQuery+eventPartQuery+eventStatusQuery)
	return err
}

func (e *EventsDB) DeleteHardEventByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event where id=$1", id)
	return err
}

func prepareEventUpdateQuery(req Event) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.RegistrationRequired != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_required=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_status=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("audience=$%d", len(updateStrings)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("slug=$%d", len(updateStrings)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("logo=$%d", len(updateStrings)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
	}
	if req.StartsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("starts_on=$%d", len(updateStrings)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("ends_on=$%d", len(updateStrings)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("date_confirmed=$%d", len(updateStrings)+1))
		args = append(args, *req.DateConfirmed)
	}
	if req.ArchiveLink != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("archive_link=$%d", len(updateStrings)+1))
		args = append(args, *req.ArchiveLink)
	}
	if req.Published != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("published=$%d", len(updateStrings)+1))
		args = append(args, *req.Published)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventCreateQuery(req Event) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.RegistrationRequired != nil {
		createStrings = append(createStrings, "registration_required")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		createStrings = append(createStrings, "registration_status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		createStrings = append(createStrings, "audience")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		createStrings = append(createStrings, "slug")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil && *req.Logo != "" {
		createStrings = append(createStrings, "logo")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil && *req.Content != "" {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.Deleted != nil {
		createStrings = append(createStrings, "deleted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Deleted)
	}
	if req.StartsOn != nil {
		createStrings = append(createStrings, "starts_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		createStrings = append(createStrings, "ends_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		createStrings = append(createStrings, "date_confirmed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.DateConfirmed)
	}
	if req.ArchiveLink != nil {
		createStrings = append(createStrings, "archive_link")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ArchiveLink)
	}
	if req.Published != nil {
		createStrings = append(createStrings, "published")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Published)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func buildAndGetWhereEventQuery(slug string) string {

	var whereString strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	whereCondition.WriteString(" deleted=false")

	// WHERE query generation based on parameters
	if slug != "" {
		whereCondition.WriteString(fmt.Sprintf(" AND slug='%s'", slug))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String()
}
