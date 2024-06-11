package repo

import (
	"context"
	"fmt"
	"strings"
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
	toUpdate, toUpdateArgs := prepareItemUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE item SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewItem(ctx context.Context, req Item) error {
	createString, numString, createQueryArgs := prepareItemCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO item (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteItemByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from item where id=$1", id)
	return err
}

func prepareItemUpdateQuery(req Item) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.StartDate != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("start_date=$%d", len(updateStrings)+1))
		args = append(args, *req.StartDate)
	}
	if req.Duration != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("duration=$%d", len(updateStrings)+1))
		args = append(args, *req.Duration)
	}
	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.OriginalLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("original_language=$%d", len(updateStrings)+1))
		args = append(args, *req.OriginalLanguage)
	}
	if req.Translated != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("translated=$%d", len(updateStrings)+1))
		args = append(args, *req.Translated)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareItemCreateQuery(req Item) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.StartDate != nil {
		createStrings = append(createStrings, "start_date")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.StartDate)
	}
	if req.Duration != nil {
		createStrings = append(createStrings, "duration")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Duration)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Content != nil {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.OriginalLanguage != nil {
		createStrings = append(createStrings, "original_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.OriginalLanguage)
	}
	if req.Translated != nil {
		createStrings = append(createStrings, "translated")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Translated)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
