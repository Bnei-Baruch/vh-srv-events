package repo

import (
	"context"
	"fmt"
	"strings"

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
	toUpdate, toUpdateArgs := preparePartOptionUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE participation_option SET %s WHERE name='%s'`, toUpdate, name), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewParticipantOption(ctx context.Context, req ParticipantOption) error {
	createString, numString, createQueryArgs := preparePartOptionCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO participation_option (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteParticipantOptionByName(ctx context.Context, name string) error {
	_, err := e.Exec(ctx, "delete from participation_option where name=$1", name)
	return err
}

func preparePartOptionCreateQuery(req ParticipantOption) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Content != nil {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.Description != nil {
		createStrings = append(createStrings, "description")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Description)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func preparePartOptionUpdateQuery(req ParticipantOption) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("description=$%d", len(updateStrings)+1))
		args = append(args, *req.Description)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
