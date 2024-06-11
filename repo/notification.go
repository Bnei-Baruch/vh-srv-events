package repo

import (
	"context"
	"fmt"

	"gitlab.bbdev.team/vh/vh-srv-events/pkg/utils"
)

type Notification struct {
	Language      *string `json:"language" validate:"required"`
	EventSlug     *string `json:"event_slug" validate:"required"`
	EventTemplate *string `json:"email_template" validate:"required"`
	FromEmail     *string `json:"from_email"`
	FromName      *string `json:"from_name"`
}

func (e *EventsDB) FetchUsersAndSendEmail(ctx context.Context, s Notification) error {
	rows, err := e.Query(ctx,
		`SELECT DISTINCT p.email, p.first_name, p.last_name from participant as p, participation_status as ps, event as e
		WHERE ps.event_id = (SELECT id FROM event WHERE slug = '$1') AND
		ps.participant_id = p.id AND
		p.email_language = '$2';
	`, *s.EventSlug, *s.Language)
	if err != nil {
		return fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d Part
		err := rows.Scan(&d.Email, &d.FirstName, &d.LastName)
		if err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}

		emailErr := utils.SendEmail(ctx, s.FromName, s.FromEmail, *s.EventTemplate, *d.Email, *d.FirstName, *d.LastName)
		if emailErr != nil {
			return emailErr
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows.Err: %w", err)
	}

	return nil
}
