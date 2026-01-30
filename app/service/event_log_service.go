package service

import (
	"context"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// createEventLog persists a prepared EventLog entry using the database layer.
func (s *Service) createEventLog(ctx context.Context, log *ds.EventLog) error {
	ctx, span := s.tracer.Start(ctx, "createEventLog")
	defer span.End()

	return s.db.CreateEventLog(ctx, log)
}

// LogEntityCreated records an event related to entity creation.
//
// If the entity requires moderation, a hidden "entity_submitted" event is created.
// If the entity is already approved (e.g. created by an admin or trusted user),
// a public "entity_created" event is recorded immediately.
//
// Public-facing feeds should only display the final "entity_created" event.
func (s *Service) LogEntityCreated(ctx context.Context, e *ds.Entity) error {
	ctx, span := s.tracer.Start(ctx, "LogEntityCreated")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(e.OwnerID),
		Type:     ds.EventLogEntitySubmitted,
		EntityID: app.Pointer(e.ID),
		IsPublic: false,
	}

	if e.Status == ds.EntityStatusApproved {
		log.Type = ds.EventLogEntityAdded
		log.IsPublic = true
	}

	return s.createEventLog(ctx, log)
}

// LogEntityUpdated records a public-facing entity update event.
func (s *Service) LogEntityUpdated(ctx context.Context, e *ds.Entity) error {
	ctx, span := s.tracer.Start(ctx, "LogEntityUpdated")
	defer span.End()

	// TODO: attach change request reference (ChangeRequestID)
	log := &ds.EventLog{
		UserID:   app.Pointer(e.OwnerID),
		Type:     ds.EventLogEntityUpdated,
		EntityID: app.Pointer(e.ID),
		IsPublic: true,
	}

	return s.createEventLog(ctx, log)
}

// LogUserRegistered records the creation of a user account.
func (s *Service) LogUserRegistered(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogUserRegistered")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserAccountCreated,
		IsPublic: false,
	}

	err := s.createEventLog(ctx, log)
	if err != nil {
		return err
	}

	log2 := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserAccountActivated,
		IsPublic: false,
	}

	return s.createEventLog(ctx, log2)
}

// LogEmailConfirmed records successful email confirmation for a user.
func (s *Service) LogEmailConfirmed(ctx context.Context, email string, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogEmailConfirmed")
	defer span.End()

	log := &ds.EventLog{
		UserID: app.Pointer(userID),
		Type:   ds.EventLogUserEmailConfirmed,
		Meta: map[string]any{
			"email": email,
		},
		IsPublic: false,
	}

	err := s.createEventLog(ctx, log)
	if err != nil {
		return err
	}

	log2 := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserAccountActivated,
		IsPublic: true,
	}

	return s.createEventLog(ctx, log2)
}

// LogPasswordResetRequest records that a user has requested a password reset.
func (s *Service) LogPasswordResetRequest(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogPasswordResetRequest")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserRequestPasswordReset,
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogPasswordChangedByResetRequest records a password change performed.
func (s *Service) LogPasswordChangedByResetRequest(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogPasswordChangedByResetRequest")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserPasswordChangedByResetRequest,
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogPasswordChanged records a password change initiated by the user
// while authenticated.
func (s *Service) LogPasswordChanged(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogPasswordChanged")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserPasswordChanged,
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogEmailChangeRequested records that a user has initiated an email change flow.
func (s *Service) LogEmailChangeRequested(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogEmailChangedRequested")
	defer span.End()

	log := &ds.EventLog{
		UserID:   app.Pointer(userID),
		Type:     ds.EventLogUserEmailChangeRequested,
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogEmailChanged records a completed email address change for a user.
func (s *Service) LogEmailChanged(ctx context.Context, userID ds.ID, oldEmail, newEmail string) error {
	ctx, span := s.tracer.Start(ctx, "LogEmailChanged")
	defer span.End()

	log := &ds.EventLog{
		UserID: app.Pointer(userID),
		Type:   ds.EventLogUserEmailChanged,
		Meta: map[string]any{
			"old_email": oldEmail,
			"new_email": newEmail,
		},
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogUsernameChanged records a username change performed by the user.
func (s *Service) LogUsernameChanged(ctx context.Context, userID ds.ID, oldName, newName string) error {
	ctx, span := s.tracer.Start(ctx, "LogEmailChanged")
	defer span.End()

	log := &ds.EventLog{
		UserID: app.Pointer(userID),
		Type:   ds.EventLogUserUsernameChanged,
		Meta: map[string]any{
			"old_username": oldName,
			"new_username": newName,
		},
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}
