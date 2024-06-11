package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type EventPartOptionResponse struct {
	ID                  *int       `json:"id,omitempty" db:"id"`
	EventID             *int       `json:"event_id,omitempty" db:"event_id"`
	ParticipationOption *string    `json:"participation_option,omitempty" db:"participation_option"`
	Deleted             *bool      `json:"deleted,omitempty" db:"deleted"`
	CreatedAt           *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type EventPartOption struct {
	EventID             *int    `json:"event_id" db:"event_id" validate:"required"`
	ParticipationOption *string `json:"participation_option" db:"participation_option" validate:"required"`
	Deleted             *bool   `json:"deleted" db:"deleted"`
}

func (e *EventsDB) GetEventPartOptionByID(ctx context.Context, id string) (*EventPartOptionResponse, error) {
	u := EventPartOptionResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option where id = $1`, id).Scan(
		&u.ID,
		&u.EventID,
		&u.ParticipationOption,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllEventPartOption(ctx context.Context, skip int, limit int) ([]EventPartOptionResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []EventPartOptionResponse{}
	for rows.Next() {
		var d EventPartOptionResponse
		err := rows.Scan(&d.ID, &d.EventID, &d.ParticipationOption, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateEventPartOptionByID(ctx context.Context, req EventPartOption, id string) error {
	toUpdate, toUpdateArgs := prepareEventPartOptionUpdateQuery(req)

	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE event_participation_option SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewEventPartOption(ctx context.Context, req EventPartOption) error {
	createString, numString, createQueryArgs := prepareEventPartOptionCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return common.ErrInvalidValues
	}

	_, err := e.Exec(ctx, fmt.Sprintf(`INSERT INTO event_participation_option (%s) VALUES (%s)`, createString, numString), createQueryArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	return nil
}

func (e *EventsDB) DeleteEventPartOptionByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from event_participation_option where id=$1", id)
	return err
}

func prepareEventPartOptionUpdateQuery(req EventPartOption) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
	}
	if req.ParticipationOption != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participation_option=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventPartOptionCreateQuery(req EventPartOption) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.ParticipationOption != nil {
		createStrings = append(createStrings, "participation_option")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipationOption)
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
