package repo

import (
	"context"
	"fmt"
	"strings"
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
	toUpdate, toUpdateArgs := prepareEventItemUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE event_item SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewEventItem(ctx context.Context, req EventItem) error {
	createString, numString, createQueryArgs := prepareEventItemCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO event_item (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventItemByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event_item where id=$1", id)
	return err
}

func prepareEventItemUpdateQuery(req EventItem) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
	}
	if req.ItemID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("item_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ItemID)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventItemCreateQuery(req EventItem) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.ItemID != nil {
		createStrings = append(createStrings, "item_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ItemID)
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
