package repo

import (
	"context"
	"fmt"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type Platform struct {
	Name *string `json:"Name" db:"Name" validate:"required"`
}

func (e *EventsDB) GetPlatformByName(ctx context.Context, name string) (*Platform, error) {
	u := Platform{}
	if err := e.QueryRow(ctx, `select name from platform where name = $1`, name).
		Scan(&u.Name); err != nil {
		return nil, err
	}
	return &u, nil
}

func (e *EventsDB) GetAllPlatform(ctx context.Context, skip int, limit int) ([]Platform, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select name from platform LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []Platform{}
	for rows.Next() {
		var d Platform
		err := rows.Scan(&d.Name)
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

func (e *EventsDB) UpdatePlatformByName(ctx context.Context, req Platform, name string) error {
	if req.Name == nil {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, `UPDATE platform SET name=$1 WHERE name=$2`, req.Name, name)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewPlatform(ctx context.Context, req Platform) error {
	_, err := e.Exec(ctx, `INSERT INTO platform (name) VALUES ($1)`, *req.Name)
	return err
}

func (e *EventsDB) DeletePlatformByName(ctx context.Context, name string) error {
	_, err := e.Exec(ctx, "delete from platform where name=$1", name)
	return err
}
