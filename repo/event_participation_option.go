package repo

import (
	"context"
	"fmt"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type EventPartOptionResponse struct {
	ID                  *int       `json:"id,omitempty" db:"id"`
	EventID             *int       `json:"event_id,omitempty" db:"event_id"`
	ParticipationOption *string    `json:"participation_option,omitempty" db:"participation_option"`
	Deleted             *bool      `json:"deleted,omitempty" db:"deleted"`
	CreatedAt           *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type EventPartOption struct {
	EventID             *int    `json:"event_id" db:"event_id" validate:"required"`
	ParticipationOption *string `json:"participation_option" db:"participation_option" validate:"required"`
	Deleted             *bool   `json:"deleted" db:"deleted"`
}

func (e *EventsDB) GetEventPartOptionByID(ctx context.Context, id string) (*EventPartOptionResponse, error) {
	u := EventPartOptionResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option where id = $1`, id).Scan(
		&u.ID,
		&u.EventID,
		&u.ParticipationOption,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllEventPartOption(ctx context.Context, skip int, limit int) ([]EventPartOptionResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []EventPartOptionResponse{}
	for rows.Next() {
		var d EventPartOptionResponse
		err := rows.Scan(&d.ID, &d.EventID, &d.ParticipationOption, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateEventPartOptionByID(ctx context.Context, req EventPartOption, id string) error {
	c := eventPartOptionUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("event_participation_option", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewEventPartOption(ctx context.Context, req EventPartOption) error {
	c := eventPartOptionCreateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("event_participation_option")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventPartOptionByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event_participation_option where id=$1", id)
	return err
}

func eventPartOptionUpdateColumns(req EventPartOption) columns {
	var c columns
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "participation_option", req.ParticipationOption)
	setIf(&c, "deleted", req.Deleted)
	return c
}

func eventPartOptionCreateColumns(req EventPartOption) columns {
	var c columns
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "participation_option", req.ParticipationOption)
	setIf(&c, "deleted", req.Deleted)
	return c
}
