package repo

import (
	"context"
	"fmt"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type Audience struct {
	Name        *string `json:"Name" db:"Name" validate:"required"`
	Description *string `json:"description,omitempty" db:"description"`
}

func (e *EventsDB) GetAudienceByName(ctx context.Context, name string) (*Audience, error) {
	u := Audience{}
	if err := e.QueryRow(ctx, `select name, description from audience where name = $1`, name).
		Scan(&u.Name, &u.Description); err != nil {
		return nil, err
	}
	return &u, nil
}

func (e *EventsDB) GetAllAudience(ctx context.Context, skip int, limit int) ([]Audience, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select name, description from audience LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []Audience{}
	for rows.Next() {
		var d Audience
		err := rows.Scan(&d.Name, &d.Description)
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

func (e *EventsDB) UpdateAudienceByName(ctx context.Context, req Audience, name string) error {

	c := audienceUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.update("audience", "name", name)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewAudience(ctx context.Context, req Audience) error {
	c := audienceCreateColumns(req)
	if !c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("audience")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteAudienceByName(ctx context.Context, name string) error {
	_, err := e.Exec(ctx, "delete from audience where name=$1", name)
	return err
}

func audienceUpdateColumns(req Audience) columns {
	var c columns
	setIf(&c, "name", req.Name)
	setIf(&c, "description", req.Description)
	return c
}

func audienceCreateColumns(req Audience) columns {
	var c columns
	setIf(&c, "name", req.Name)
	setIf(&c, "description", req.Description)
	return c
}
