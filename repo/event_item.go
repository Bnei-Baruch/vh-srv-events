package repo

import (
	"context"
	"fmt"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type EventItemResponse struct {
	ID        *int       `json:"id" db:"id"`
	EventID   *int       `json:"event_id" db:"event_id"`
	ItemID    *int       `json:"item_id" db:"item_id"`
	Deleted   *bool      `json:"deleted" db:"deleted"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type EventItem struct {
	EventID *int  `json:"event_id" db:"event_id" validate:"required"`
	ItemID  *int  `json:"item_id" db:"item_id" validate:"required"`
	Deleted *bool `json:"deleted" db:"deleted"`
}

func (e *EventsDB) GetEventItemByID(ctx context.Context, id string) (*EventItemResponse, error) {
	u := EventItemResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	event_id,
	item_id,
	deleted,
	created_at,
	updated_at 
	from event_item where id = $1`, id).Scan(
		&u.ID,
		&u.EventID,
		&u.ItemID,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllEventItem(ctx context.Context, skip int, limit int) ([]EventItemResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	event_id,
	item_id,
	deleted,
	created_at,
	updated_at 
	from event_item LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []EventItemResponse{}
	for rows.Next() {
		var d EventItemResponse
		err := rows.Scan(&d.ID, &d.EventID, &d.ItemID, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateEventItemByID(ctx context.Context, req EventItem, id string) error {
	c := eventItemUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("event_item", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewEventItem(ctx context.Context, req EventItem) error {
	c := eventItemCreateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("event_item")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventItemByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event_item where id=$1", id)
	return err
}

func eventItemUpdateColumns(req EventItem) columns {
	var c columns
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "item_id", req.ItemID)
	return c
}

func eventItemCreateColumns(req EventItem) columns {
	var c columns
	setIf(&c, "event_id", req.EventID)
	setIf(&c, "item_id", req.ItemID)
	setIf(&c, "deleted", req.Deleted)
	return c
}
