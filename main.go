package main

import (
	"log/slog"

	"github.com/niksteff/go-eventSourcing/internal/events"
	account "github.com/niksteff/go-eventSourcing/internal/repository"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	repo := account.NewEventRepository()
	history := repo.GetHistoryForAccount("9fd16a83-f0b1-4301-9d63-f3f151ae2dbd")
	aggregate := events.NewFrequentFlierAccountFromHistory(history)

	slog.Info("Before recording flight", slog.String("account", aggregate.String()))
	aggregate.RecordFlightTaken(1000, 3)
	slog.Info("After recording flight", slog.String("account", aggregate.String()))
}
