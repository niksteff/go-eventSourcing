package account

import (
	"github.com/google/uuid"
	"github.com/niksteff/go-eventSourcing/internal/events"
)

type EventRepository struct {
	db map[string][]any // the simulated database of a relation accountId to events
}

func NewEventRepository() *EventRepository {
	r := EventRepository{}
	r.populate()

	return &r
}

func (r *EventRepository) GetHistoryForAccount(accountId string) <-chan any {
	h := make(chan any)

	go func() {
		defer close(h)

		events, ok := r.db[accountId]
		if !ok {
			return
		}

		for _, e := range events {
			h <- e
		}
	}()

	return h
}

func (r *EventRepository) populate() {
	if r.db == nil {
		r.db = make(map[string][]any)
	}

	r.db["9fd16a83-f0b1-4301-9d63-f3f151ae2dbd"] = []any{
		events.FrequentFlierAccountCreated{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			AccountId:         "9fd16a83-f0b1-4301-9d63-f3f151ae2dbd",
			OpeningMiles:      10000,
			OpeningTierPoints: 0,
		},
		events.StatusMatched{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			NewStatus: events.StatusSilver,
		},
		events.FlightTaken{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			MilesAdded:      2525,
			TierPointsAdded: 5,
		},
		events.FlightTaken{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			MilesAdded:      2512,
			TierPointsAdded: 5,
		},
		events.FlightTaken{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			MilesAdded:      5600,
			TierPointsAdded: 5,
		},
		events.FlightTaken{
			Event: events.Event{
				Id: uuid.NewString(),
			},
			MilesAdded:      3000,
			TierPointsAdded: 3,
		},
	}
}
