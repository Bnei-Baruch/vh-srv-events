package repo

import (
	"context"
	"fmt"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type ItemResponse struct {
	ID               *int                    `json:"id" db:"id"`
	StartDate        *time.Time              `json:"start_date" db:"start_date"`
	Duration         *int                    `json:"duration" db:"duration"`
	Name             *string                 `json:"name" db:"name"`
	Content          *map[string]interface{} `json:"content,omitempty" db:"content"`
	OriginalLanguage *string                 `json:"original_language" db:"original_language"`
	Translated       *bool                   `json:"translated" db:"translated"`
	CreatedAt        *time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        *time.Time              `json:"updated_at" db:"updated_at"`
}

type Item struct {
	StartDate        *time.Time `json:"start_date" db:"start_date" validate:"required"`
	Duration         *int       `json:"duration" db:"duration" validate:"required"`
	Name             *string    `json:"name" db:"name" validate:"required"`
	Content          *string    `json:"content,omitempty" db:"content"`
	OriginalLanguage *string    `json:"original_language" db:"original_language" validate:"required"`
	Translated       *bool      `json:"translated" db:"translated" validate:"required"`
}

func (e *EventsDB) GetItemByID(ctx context.Context, id string) (*ItemResponse, error) {
	u := ItemResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	start_date,
	duration,
	name,
	original_language,
	translated,
	created_at,
	updated_at 
	from item where id = $1`, id).Scan(
		&u.ID,
		&u.StartDate,
		&u.Duration,
		&u.Name,
		&u.OriginalLanguage,
		&u.Translated,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllItem(ctx context.Context, skip int, limit int) ([]ItemResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	start_date,
	duration,
	name,
	original_language,
	translated,
	created_at,
	updated_at 
	from item LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []ItemResponse{}
	for rows.Next() {
		var d ItemResponse
		err := rows.Scan(&d.ID, &d.StartDate, &d.Duration, &d.Name, &d.OriginalLanguage, &d.Translated, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateItemByID(ctx context.Context, req Item, id string) error {
	c := itemUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("item", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewItem(ctx context.Context, req Item) error {
	c := itemCreateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("item")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteItemByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from item where id=$1", id)
	return err
}

func itemUpdateColumns(req Item) columns {
	var c columns
	setIf(&c, "start_date", req.StartDate)
	setIf(&c, "duration", req.Duration)
	setIf(&c, "name", req.Name)
	setIf(&c, "original_language", req.OriginalLanguage)
	setIf(&c, "translated", req.Translated)
	return c
}

func itemCreateColumns(req Item) columns {
	var c columns
	setIf(&c, "start_date", req.StartDate)
	setIf(&c, "duration", req.Duration)
	setIf(&c, "name", req.Name)
	setIf(&c, "content", req.Content)
	setIf(&c, "original_language", req.OriginalLanguage)
	setIf(&c, "translated", req.Translated)
	return c
}
