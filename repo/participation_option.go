package repo

import (
	"context"
	"fmt"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type ParticipantOptionResponse struct {
	Name        *string                 `json:"name" db:"name"`
	Description *string                 `json:"description,omitempty" db:"description"`
	Content     *map[string]interface{} `json:"content,omitempty" db:"content"`
}

type ParticipantOption struct {
	Name        *string `json:"name" db:"Name" validate:"required"`
	Description *string `json:"description,omitempty" db:"description"`
	Content     *string `json:"content,omitempty" db:"content"`
}

func (e *EventsDB) GetParticipantOptionByName(ctx context.Context, name string) (*ParticipantOptionResponse, error) {
	u := ParticipantOptionResponse{}
	if err := e.QueryRow(ctx, `select name, description, content from participation_option where name = $1`, name).
		Scan(&u.Name, &u.Description, &u.Content); err != nil {
		return nil, err
	}
	return &u, nil
}

func (e *EventsDB) GetAllParticipantOption(ctx context.Context, skip int, limit int) ([]ParticipantOptionResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select name, description, content from participation_option LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []ParticipantOptionResponse{}
	for rows.Next() {
		var d ParticipantOptionResponse
		err := rows.Scan(&d.Name, &d.Description, &d.Content)
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

func (e *EventsDB) UpdateParticipantOptionByName(ctx context.Context, req ParticipantOption, name string) error {
	c := partOptionUpdateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.update("participation_option", "name", name)

	updateRes, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewParticipantOption(ctx context.Context, req ParticipantOption) error {
	c := partOptionCreateColumns(req)
	if c.empty() {
		return common.ErrInvalidValues
	}
	query, args := c.insert("participation_option")

	_, err := e.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteParticipantOptionByName(ctx context.Context, name string) error {
	_, err := e.Exec(ctx, "delete from participation_option where name=$1", name)
	return err
}

func partOptionCreateColumns(req ParticipantOption) columns {
	var c columns
	setIf(&c, "content", req.Content)
	setIf(&c, "description", req.Description)
	setIf(&c, "name", req.Name)
	return c
}

func partOptionUpdateColumns(req ParticipantOption) columns {
	var c columns
	setIf(&c, "name", req.Name)
	setIf(&c, "description", req.Description)
	setIf(&c, "content", req.Content)
	return c
}
