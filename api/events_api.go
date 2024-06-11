package api

import "gitlab.bbdev.team/vh/vh-srv-events/repo"

type EventsAPI struct {
	repo repo.EventsRepository
}

func NewEventsAPI(db repo.EventsRepository) *EventsAPI {
	return &EventsAPI{
		repo: db,
	}
}
