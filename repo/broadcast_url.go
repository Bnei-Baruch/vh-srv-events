package repo

import (
	"context"
	"fmt"
	"strings"
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
	toUpdate, toUpdateArgs := prepareURLUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE broadcast_url SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
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

func prepareURLUpdateQuery(req BroadcastURL) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.URL != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("url=$%d", len(updateStrings)+1))
		args = append(args, *req.URL)
	}
	if req.Platform != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("platform=$%d", len(updateStrings)+1))
		args = append(args, *req.Platform)
	}
	if req.Language != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("language=$%d", len(updateStrings)+1))
		args = append(args, *req.Language)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
