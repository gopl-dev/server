package factory

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app/ds"
	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
)

// NewUserActivityLog ...
func (f *Factory) NewUserActivityLog(overrideOpt ...ds.UserActivityLog) (m *ds.UserActivityLog) {
	m = &ds.UserActivityLog{
		ID:         ds.NilID,
		UserID:     ds.NilID,
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
func (f *Factory) CreateUserActivityLog(overrideOpt ...ds.UserActivityLog) (m *ds.UserActivityLog, err error) {
	m = f.NewUserActivityLog(overrideOpt...)
	err = f.repo.CreateUserActivityLog(context.Background(), m)

	return
}
