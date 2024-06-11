package repo

import (
	"context"
	"fmt"
)

type PartOptionAndCount struct {
	ParticipationOption *string `json:"participation_option"`
	Count               *int    `json:"count"`
}

func (e *EventsDB) FetchTotalParticipantByOptionAndGroupBy(ctx context.Context, eventID string) ([]PartOptionAndCount, error) {
	var eventIDQuery string
	if eventID != "" {
		eventIDQuery = ` WHERE event_id=` + eventID
	}

	rows, err := e.Query(ctx, `select participation_option, count (participation_option) as qt 
		from participation_status`+eventIDQuery+
		` group by participation_option`)
	if err != nil {
		return nil, fmt.Errorf("e.Query: %w", err)
	}
	defer rows.Close()

	partOptions := []PartOptionAndCount{}
	for rows.Next() {
		var p PartOptionAndCount
		err := rows.Scan(&p.ParticipationOption, &p.Count)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		partOptions = append(partOptions, p)
	}

	return partOptions, nil
}

func (e *EventsDB) FetchTotalParticipantByOption(ctx context.Context, eventID string) (int, error) {
	var eventIDQuery string
	if eventID != "" {
		eventIDQuery = ` WHERE event_id=` + eventID
	}

	var count int
	if err := e.QueryRow(ctx, `select count (participation_option) as qt from participation_status`+eventIDQuery).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
