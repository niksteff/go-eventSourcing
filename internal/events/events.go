package events

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type Status int

const (
	StatusRed    Status = iota
	StatusSilver Status = iota
	StatusGold   Status = iota
)

type Transition interface {
}

type Event struct {
	Id string
}

func (e Event) String() string {
	return fmt.Sprintf("%T: Id=%q", e, e.Id)
}

type FrequentFlierAccountCreated struct {
	Event
	AccountId         string
	OpeningMiles      int
	OpeningTierPoints int
}

func (e FrequentFlierAccountCreated) String() string {
	return fmt.Sprintf("%T: Event=%q AccountId=%q OpeningMiles=%d OpeningTierPoints=%d", e, e.Id, e.AccountId, e.OpeningMiles, e.OpeningTierPoints)
}

type StatusMatched struct {
	Event
	NewStatus Status
}

func (e StatusMatched) String() string {
	return fmt.Sprintf("%T: Event=%q NewStatus=%v", e, e.Id, e.NewStatus)
}

type FlightTaken struct {
	Event
	MilesAdded      int
	TierPointsAdded int
}

func (e FlightTaken) String() string {
	return fmt.Sprintf("%T: Event=%q MilesAdded=%d TierPointsAdded=%d", e, e.Id, e.MilesAdded, e.TierPointsAdded)
}

type PromotedToGoldStatus struct {
	Event
}

func (e PromotedToGoldStatus) String() string {
	return fmt.Sprintf("%T: Event=%q Promotion=%q", e, e.Id, "PromotedToGold")
}

// FrequentFlierAccount is an aggretate to fold the history of an account
type FrequentFlierAccount struct {
	id         string
	miles      int
	tierPoints int
	status     Status

	expectedVersion int   // the persisted version of the current aggregate, basically +1 for each transitioned event on the aggregate
	changes         []any // pending changes applied to the aggregate but not persisted, yet
}

// NewFrequentFlierAccountFromHistory folds the given events to represent the state of an account for a specific point in time based on the given events
func NewFrequentFlierAccountFromHistory(events <-chan any) *FrequentFlierAccount {
	state := &FrequentFlierAccount{}

	for event := range events {
		state.transition(event)
		state.expectedVersion++
	}

	return state
}

// trackChange will transition the accounts state to affect changes of the given event, the events are not persisted yet
func (state *FrequentFlierAccount) trackChange(event any) {
	state.changes = append(state.changes, event)
	state.transition(event)
}

func (state *FrequentFlierAccount) transition(event any) {
	switch e := event.(type) {

	case FrequentFlierAccountCreated:
		slog.Debug("transition", slog.String("Event", e.String()))

		state.id = e.AccountId
		state.miles = e.OpeningMiles
		state.tierPoints = e.OpeningTierPoints
		state.status = StatusRed

	case StatusMatched:
		slog.Debug("transition", slog.String("Event", e.String()))

		state.status = e.NewStatus

	case FlightTaken:
		slog.Debug("transition", slog.String("Event", e.String()))

		state.miles = state.miles + e.MilesAdded
		state.tierPoints = state.tierPoints + e.TierPointsAdded

	case PromotedToGoldStatus:
		slog.Debug("transition", slog.String("Event", e.String()))

		state.status = StatusGold

	default:
		slog.Error("dropping unknown event", slog.Any("Event", e))
	}
}

func (a FrequentFlierAccount) String() string {
	return fmt.Sprintf("FrequentFlierAccount=%q, Miles=%d, TierPoints=%d, Status=%v ExpectedVersion=%d PendingChanges=%d", a.id, a.miles, a.tierPoints, a.status, a.expectedVersion, len(a.changes))
}

// RecordFlightTaken is used to record the fact that a customer has taken a flight
// which should be attached to this frequent flier account. The number of miles and
// tier points which apply are calculated externally.
//
// If recording this flight takes the account over a status boundary, it will
// automatically upgrade the account to the new status level.
func (self *FrequentFlierAccount) RecordFlightTaken(miles int, tierPoints int) {
	// Obviously we should be doing some validation here...

	// issue a flight taken event
	self.trackChange(FlightTaken{
		Event: Event{
			Id: uuid.NewString(),
		},
		MilesAdded:      miles,
		TierPointsAdded: tierPoints,
	})

	if self.tierPoints > 20 && self.status != StatusGold {
		// issue another event when we reached gold state
		self.trackChange(PromotedToGoldStatus{
			Event: Event{
				Id: uuid.NewString(),
			},
		})
	}
}
