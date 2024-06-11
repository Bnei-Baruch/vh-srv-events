package repo

import (
	"context"
	"fmt"
	"strings"

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
	toUpdate, toUpdateArgs := prepareAudienceUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE audience SET %s WHERE name='%s'`, toUpdate, name), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewAudience(ctx context.Context, req Audience) error {
	createString, numString, createQueryArgs := prepareAudienceCreateQuery(req)
	if len(createQueryArgs) != 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO audience (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteAudienceByName(ctx context.Context, name string) error {
	_, err := e.Exec(ctx, "delete from audience where name=$1", name)
	return err
}

func prepareAudienceUpdateQuery(req Audience) (string, []interface{}) {
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

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareAudienceCreateQuery(req Audience) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		createStrings = append(createStrings, "description")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Description)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
