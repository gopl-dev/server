package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
)

// NewUserActivityLog ...
func (f *Factory) NewUserActivityLog(overrideOpt ...ds.UserActivityLog) (m *ds.UserActivityLog) {
	m = &ds.UserActivityLog{
		ID:         0,
		UserID:     0,
		ActionType: useractivity.UserRegistered,
		IsPublic:   false,
		CreatedAt:  time.Now(),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateUserActivityLog ...
func (f *Factory) CreateUserActivityLog(t *testing.T, overrideOpt ...ds.UserActivityLog) (m *ds.UserActivityLog) {
	t.Helper()

	m = f.NewUserActivityLog(overrideOpt...)

	err := f.repo.CreateUserActivityLog(context.Background(), m)
	checkErr(t, err)

	return
}
