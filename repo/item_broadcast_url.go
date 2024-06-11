package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type ItemBroadcastURLResponse struct {
	ID             *int       `json:"id" db:"id"`
	ItemID         *int       `json:"item_id" db:"item_id"`
	BoradcastURLID *int       `json:"broadcast_url_id" db:"broadcast_url_id"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

type ItemBroadcastURL struct {
	ItemID         *int `json:"item_id" db:"item_id" validate:"required"`
	BoradcastURLID *int `json:"broadcast_url_id" db:"broadcast_url_id" validate:"required"`
}

func (e *EventsDB) GetItemBroadcastURLByID(ctx context.Context, id string) (*ItemBroadcastURLResponse, error) {
	u := ItemBroadcastURLResponse{}
	if err := e.QueryRow(ctx, `select 
	id,
	item_id,
	broadcast_url_id,
	created_at,
	updated_at 
	from item_broadcast_url where id = $1`, id).Scan(
		&u.ID,
		&u.ItemID,
		&u.BoradcastURLID,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (e *EventsDB) GetAllItemBroadcastURL(ctx context.Context, skip int, limit int) ([]ItemBroadcastURLResponse, error) {
	rows, err := e.Query(ctx, fmt.Sprintf(`select 
	id,
	item_id,
	broadcast_url_id,
	created_at,
	updated_at 
	from item_broadcast_url LIMIT %d OFFSET %d`, limit, skip))
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	u := []ItemBroadcastURLResponse{}
	for rows.Next() {
		var d ItemBroadcastURLResponse
		err := rows.Scan(&d.ID, &d.ItemID, &d.BoradcastURLID, &d.CreatedAt, &d.UpdatedAt)
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

func (e *EventsDB) UpdateItemBroadcastURLByID(ctx context.Context, req ItemBroadcastURL, id string) error {
	toUpdate, toUpdateArgs := prepareItemBroadcastURLUpdateQuery(req)
	if len(toUpdateArgs) == 0 {
		return common.ErrInvalidValues
	}

	updateRes, err := e.Exec(ctx, fmt.Sprintf(`UPDATE item_broadcast_url SET %s WHERE id=%s`, toUpdate, id), toUpdateArgs...)
	if err != nil {
		return fmt.Errorf("e.Exec: %w", err)
	}

	if updateRes.RowsAffected() == 0 {
		return common.ErrNoRowsAffected
	}

	return nil
}

func (e *EventsDB) CreateNewItemBroadcastURL(ctx context.Context, req ItemBroadcastURL) error {
	_, err := e.Exec(ctx, `INSERT INTO item_broadcast_url (item_id,broadcast_url_id) VALUES ($1,$2)`,
		*req.ItemID, *req.BoradcastURLID)

	return err
}

func (e *EventsDB) DeleteItemBroadcastURLByID(ctx context.Context, id string) error {
	_, err := e.Exec(ctx, "delete from item_broadcast_url where id=$1", id)
	return err
}

func prepareItemBroadcastURLUpdateQuery(req ItemBroadcastURL) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.BoradcastURLID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("broadcast_url_id=$%d", len(updateStrings)+1))
		args = append(args, *req.BoradcastURLID)
	}
	if req.ItemID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("item_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ItemID)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
