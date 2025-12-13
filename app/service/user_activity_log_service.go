package service

import (
	"context"
	"errors"
	"time"

	"github.com/gopl-dev/server/app/ds"
	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
	"github.com/gopl-dev/server/app/repo"
)

// createUserActivityLog creates a new user activity log entry.
func (s *Service) createUserActivityLog(ctx context.Context, log *ds.UserActivityLog) error {
	ctx, span := s.tracer.Start(ctx, "CreateUserActivityLog")
	defer span.End()

	return s.db.CreateUserActivityLog(ctx, log)
}

// LogUserRegistered creates a private activity log to record when a new user signs up.
// The log is initially private and is made public only after email confirmation.
func (s *Service) LogUserRegistered(ctx context.Context, userID int64) error {
	ctx, span := s.tracer.Start(ctx, "LogUserRegistered")
	defer span.End()

	return s.createUserActivityLog(ctx, &ds.UserActivityLog{
		UserID:     userID,
		ActionType: useractivity.UserRegistered,
		IsPublic:   false,
		CreatedAt:  time.Now(),
	})
}

// LogEmailConfirmed marks the original user registration event as complete by making it public.
func (s *Service) LogEmailConfirmed(ctx context.Context, userID int64) error {
	ctx, span := s.tracer.Start(ctx, "LogEmailConfirmed")
	defer span.End()

	// Once user confirmed email, we need to make his registration event made public
	log, err := s.db.FindUserActivityLogByUserAndType(ctx, userID, useractivity.UserRegistered)
	if errors.Is(err, repo.ErrActivityLogNotFound) {
		return nil
	}

	// We do not log a "email confirmed" event. When a user registers,
	// the activity log is created but kept private, in case the user does not
	// complete the sign-up.
	// When the user confirms their email, we make the original registration log public
	// to mark the registration as complete.
	return s.db.UpdateUserActivityLogPublic(ctx, log.ID)
}
