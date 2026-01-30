package factory

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewEventLog creates a new EventLog instance populated with sensible default
// values for testing and seeding purposes.
func (f *Factory) NewEventLog(overrideOpt ...ds.EventLog) (m *ds.EventLog) {
	m = &ds.EventLog{
		ID:        ds.NewID(),
		UserID:    nil,
		Type:      random.Element(ds.EventLogTypes),
		EntityID:  nil,
		Message:   "",
		Meta:      nil,
		IsPublic:  random.Bool(),
		CreatedAt: time.Now(),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateEventLog creates and persists a new EventLog using the repository.
func (f *Factory) CreateEventLog(overrideOpt ...ds.EventLog) (m *ds.EventLog, err error) {
	m = f.NewEventLog(overrideOpt...)
	err = f.repo.CreateEventLog(context.Background(), m)

	return
}
