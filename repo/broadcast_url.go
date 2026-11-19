package repo

import (
	"context"
	"fmt"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type BroadcastURLResponse struct {
	ID        *int       `json:"id" db:"id"`
	URL       *string    `json:"url" db:"url"`
	Platform  *string    `json:"platform" db:"platform"`
	Language  *string    `json:"language" db:"language"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type BroadcastURL struct {
	URL      *string `json:"url" db:"url" validate:"required"`
	Platform *string `json:"platform" db:"platform" validate:"required"`
	Language *string `json:"language" db:"language" validate:"required"`
}

func (e *EventsDB) GetURLByID(ctx context.Context, id string) (BroadcastURLResponse, error) {
	u := BroadcastURLResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	url,
	platform,
	language,
	created_at,
	updated_at 
	from broadcast_url where id = $1`, id).Scan(
		&u.ID,
		&u.URL,
		&u.Platform,
		&u.Language,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return BroadcastURLResponse{}, err
	}

	return u, nil
}

func (e *EventsDB) GetAllURL(ctx context.Context, skip int, limit int) ([]BroadcastURLResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	url,
	platform,
	language,
	created_at,
	updated_at 
	from broadcast_url LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []BroadcastURLResponse{}
	for rows.Next() {
		var d BroadcastURLResponse
		err := rows.Scan(&d.ID, &d.URL, &d.Platform, &d.Language, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateURLByID(ctx context.Context, req BroadcastURL, id string) error {
	c := urlUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	c.set("updated_at", time.Now())
	query, args := c.update("broadcast_url", "id", id)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewURL(ctx context.Context, req BroadcastURL) error {
	_, err := e.Exec(ctx,
		`INSERT INTO broadcast_url (url, platform, language) VALUES ($1, $2, $3)`,
		*req.URL,
		*req.Platform,
		*req.Language)

	return err
}

func (e *EventsDB) DeleteURLByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from broadcast_url where id=$1", id)
	return err
}

func urlUpdateColumns(req BroadcastURL) columns {
	var c columns
	setIf(&c, "url", req.URL)
	setIf(&c, "platform", req.Platform)
	setIf(&c, "language", req.Language)
	return c
}
