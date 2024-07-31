package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type EventsRepository interface {
	GetAudienceByName(ctx context.Context, name string) (*Audience, error)
	GetAllAudience(ctx context.Context, skip int, limit int) ([]Audience, error)
	UpdateAudienceByName(ctx context.Context, req Audience, name string) error
	CreateNewAudience(ctx context.Context, req Audience) error
	DeleteAudienceByName(ctx context.Context, name string) error

	GetURLByID(ctx context.Context, id string) (BroadcastURLResponse, error)
	GetAllURL(ctx context.Context, skip int, limit int) ([]BroadcastURLResponse, error)
	UpdateURLByID(ctx context.Context, req BroadcastURL, id string) error
	CreateNewURL(ctx context.Context, req BroadcastURL) error
	DeleteURLByID(ctx context.Context, id string) error

	GetItemByID(ctx context.Context, id string) (*ItemResponse, error)
	GetAllItem(ctx context.Context, skip int, limit int) ([]ItemResponse, error)
	UpdateItemByID(ctx context.Context, req Item, id string) error
	CreateNewItem(ctx context.Context, req Item) error
	DeleteItemByID(ctx context.Context, id string) error

	GetItemBroadcastURLByID(ctx context.Context, id string) (*ItemBroadcastURLResponse, error)
	GetAllItemBroadcastURL(ctx context.Context, skip int, limit int) ([]ItemBroadcastURLResponse, error)
	UpdateItemBroadcastURLByID(ctx context.Context, req ItemBroadcastURL, id string) error
	CreateNewItemBroadcastURL(ctx context.Context, req ItemBroadcastURL) error
	DeleteItemBroadcastURLByID(ctx context.Context, id string) error

	GetEventItemByID(ctx context.Context, id string) (*EventItemResponse, error)
	GetAllEventItem(ctx context.Context, skip int, limit int) ([]EventItemResponse, error)
	UpdateEventItemByID(ctx context.Context, req EventItem, id string) error
	CreateNewEventItem(ctx context.Context, req EventItem) error
	DeleteEventItemByID(ctx context.Context, id string) error

	GetEventPartOptionByID(ctx context.Context, id string) (*EventPartOptionResponse, error)
	GetAllEventPartOption(ctx context.Context, skip int, limit int) ([]EventPartOptionResponse, error)
	UpdateEventPartOptionByID(ctx context.Context, req EventPartOption, id string) error
	CreateNewEventPartOption(ctx context.Context, req EventPartOption) error
	DeleteEventPartOptionByID(ctx context.Context, id string) error

	GetEventByID(ctx context.Context, id string) (*EventResponse, error)
	GetAllEvent(ctx context.Context, skip int, limit int, slug string, email string, kcID string) ([]EventResponse, error)
	UpdateEventByID(ctx context.Context, req Event, id string) error
	CreateEvent(ctx context.Context, req Event) error
	DeleteEventByID(ctx context.Context, id string) error
	DeleteHardEventByID(ctx context.Context, id string) error

	GetPlatformByName(ctx context.Context, name string) (*Platform, error)
	GetAllPlatform(ctx context.Context, skip int, limit int) ([]Platform, error)
	UpdatePlatformByName(ctx context.Context, req Platform, name string) error
	CreateNewPlatform(ctx context.Context, req Platform) error
	DeletePlatformByName(ctx context.Context, name string) error

	GetParticipantById(ctx context.Context, id string) (*PartResponse, error)
	GetParticipantByEmail(ctx context.Context, email string) (*PartResponse, error)
	GetParticipantByKeycloakID(ctx context.Context, id string) (*PartResponse, error)
	GetAllParticipants(ctx context.Context, skip int, limit int, eventId int, eventSlug string) ([]PartResponse, error)
	UpdateParticipantByID(ctx context.Context, req Part, id string) error
	CreateNewParticipant(ctx context.Context, req Part) (int, error)
	DeleteParticipantByID(ctx context.Context, id string) error
	IsSubjectID(ctx context.Context, keycloakID, accountID string) (bool, error)

	GetParticipantOptionByName(ctx context.Context, name string) (*ParticipantOptionResponse, error)
	GetAllParticipantOption(ctx context.Context, skip int, limit int) ([]ParticipantOptionResponse, error)
	UpdateParticipantOptionByName(ctx context.Context, req ParticipantOption, name string) error
	CreateNewParticipantOption(ctx context.Context, req ParticipantOption) error
	DeleteParticipantOptionByName(ctx context.Context, name string) error

	GetParticipationStatusByID(ctx context.Context, id string) (*ParticipationStatusResponse, error)
	GetAllParticipationStatus(ctx context.Context, skip string, limit string, eventID string, keycloakID string, country string,
		email string, gender string, partOption string, firstName string, lastName string) ([]ParticipationStatusResponse, error)
	UpdateParticipationStatusByID(ctx context.Context, req ParticipationStatusStruct, id string) error
	UpdateParticipationStatusByKcIDAndEventSlug(ctx context.Context, req ParticipationStatusStruct, kcID string, eventSlug string) error
	CreateNewParticipationStatus(ctx context.Context, req PartStatusWithNotification) (int, error)
	GetTotalParticipationStatusCount(ctx context.Context, eventID string, keycloakID string, country string,
		email string, gender string, partOption string, firstName string, lastName string) (int, error)
	DeleteParticipationStatusByID(ctx context.Context, id string) error

	FetchTotalParticipantByOptionAndGroupBy(ctx context.Context, eventID string) ([]PartOptionAndCount, error)
	FetchTotalParticipantByOption(ctx context.Context, eventID string) (int, error)

	FetchUsersAndSendEmail(ctx context.Context, s Notification) error

	Close()
}

type EventsDB struct {
	*pgxpool.Pool
}

func NewEventsDB(ctx context.Context) (EventsRepository, error) {
	pool, err := pgxpool.Connect(ctx, GetDBURL())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.Connect: %w", err)
	}

	return &EventsDB{Pool: pool}, nil
}
