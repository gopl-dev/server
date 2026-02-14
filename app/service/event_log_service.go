package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// createEventLog persists a prepared EventLog entry using the database layer.
func (s *Service) createEventLog(ctx context.Context, log *ds.EventLog) error {
	ctx, span := s.tracer.Start(ctx, "createEventLog")
	defer span.End()

	return s.db.CreateEventLog(ctx, log)
}

// FilterEventLogs retrieves a paginated list of event logs matching the given filter.
func (s *Service) FilterEventLogs(ctx context.Context, f ds.EventLogsFilter) (data []ds.EventLog, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterEventLogs")
	defer span.End()

	return s.db.FilterEventLogs(ctx, f)
}

// EventLogChanges returns the changes data from an event log's metadata.
func (s *Service) EventLogChanges(ctx context.Context, id ds.ID) (changes any, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterEventLogs")
	defer span.End()

	log, err := s.db.GetEventLogByID(ctx, id)
	if err != nil {
		return
	}

	changes = log.Meta["changes"]
	return
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
		UserID:   new(e.OwnerID),
		Type:     ds.EventLogEntitySubmitted,
		EntityID: new(e.ID),
		Meta:     map[string]any{"entity_title": e.Title},
		IsPublic: false,
	}

	if e.Status == ds.EntityStatusApproved {
		log.Type = ds.EventLogEntityAdded
		log.IsPublic = true
	}

	return s.createEventLog(ctx, log)
}

// LogEntityUpdated records a public-facing entity update event.
func (s *Service) LogEntityUpdated(ctx context.Context, userID, entityID ds.ID, title, changes any) error {
	ctx, span := s.tracer.Start(ctx, "LogEntityUpdated")
	defer span.End()

	log := &ds.EventLog{
		UserID:   new(userID),
		Type:     ds.EventLogEntityUpdated,
		EntityID: new(entityID),
		Meta: map[string]any{
			"entity_title": title,
			"changes":      changes,
		},
		IsPublic: true,
	}

	return s.createEventLog(ctx, log)
}

// LogEntityRenamed records an entity rename event.
func (s *Service) LogEntityRenamed(ctx context.Context, userID, entityID ds.ID, oldTitle, newTitle any) error {
	ctx, span := s.tracer.Start(ctx, "LogEntityRenamed")
	defer span.End()

	log := &ds.EventLog{
		UserID:   new(userID),
		Type:     ds.EventLogEntityRenamed,
		EntityID: new(entityID),
		Meta: map[string]any{
			"entity_title": oldTitle,
			"new_title":    newTitle,
		},
		IsPublic: true,
	}

	return s.createEventLog(ctx, log)
}

// LogUserRegistered records the creation of a user account.
func (s *Service) LogUserRegistered(ctx context.Context, userID ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "LogUserRegistered")
	defer span.End()

	log := &ds.EventLog{
		UserID:   new(userID),
		Type:     ds.EventLogUserAccountCreated,
		IsPublic: false,
	}

	err := s.createEventLog(ctx, log)
	if err != nil {
		return err
	}

	log2 := &ds.EventLog{
		UserID:   new(userID),
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
		UserID: new(userID),
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
		UserID:   new(userID),
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
		UserID:   new(userID),
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
		UserID:   new(userID),
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
		UserID:   new(userID),
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
		UserID:   new(userID),
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
		UserID: new(userID),
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
		UserID: new(userID),
		Type:   ds.EventLogUserUsernameChanged,
		Meta: map[string]any{
			"old_username": oldName,
			"new_username": newName,
		},
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}

// LogBookApproved writes event logs for a successfully approved book.
//
// It creates two event log records:
//  1. A private log indicating that the book was approved by reviewer.
//  2. A public log for the book owner indicating that the book was added.
func (s *Service) LogBookApproved(ctx context.Context, approvedBy ds.ID, book *ds.Book) error {
	ctx, span := s.tracer.Start(ctx, "LogBookApproved")
	defer span.End()

	log := &ds.EventLog{
		UserID:   new(approvedBy),
		Type:     ds.EventLogEntityApproved,
		EntityID: new(book.ID),
		IsPublic: false,
	}
	err := s.createEventLog(ctx, log)
	if err != nil {
		return err
	}

	log2 := &ds.EventLog{
		UserID:   new(book.OwnerID),
		Type:     ds.EventLogEntityAdded,
		EntityID: new(book.ID),
		IsPublic: true,
	}
	return s.createEventLog(ctx, log2)
}

// LogBookRejected writes an event log for a rejected book.
func (s *Service) LogBookRejected(ctx context.Context, rejectedBy ds.ID, note string, book *ds.Book) error {
	ctx, span := s.tracer.Start(ctx, "LogBookApproved")
	defer span.End()

	log := &ds.EventLog{
		UserID:   new(rejectedBy),
		Type:     ds.EventLogEntityRejected,
		EntityID: new(book.ID),
		Meta: map[string]any{
			"note": note,
		},
		IsPublic: false,
	}

	return s.createEventLog(ctx, log)
}
