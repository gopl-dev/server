package service

import (
	"context"
	"maps"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// EntityChange represents the effective editable state of an entity.
type EntityChange struct {
	ID           ds.ID          `json:"id"`
	Data         map[string]any `json:"data"`
	Revision     int            `json:"revision"`
	RevisionDate *time.Time     `json:"revision_date"`
}

// GetEntityChangeState returns the effective state of an entity for the current user.
//
// If the user has no pending change request for the given entity, it returns the
// original entity data with revision 0. If a pending change request exists,
// its diff is applied on top of the provided data, and the resulting state,
// revision, and last update time are returned.
func (s *Service) GetEntityChangeState(ctx context.Context, entityID ds.ID, data ds.DataProvider) (state *EntityChange, err error) {
	ctx, span := s.tracer.Start(ctx, "GetEntityChangeState")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return nil, app.ErrUnauthorized()
	}

	req, err := s.db.FindPendingEntityChangeRequest(ctx, entityID, user.ID)
	if err != nil {
		return nil, err
	}

	// no changes were made yet, return data as is
	if req == nil {
		state = &EntityChange{
			ID:           entityID,
			Data:         data.Data(),
			Revision:     0,
			RevisionDate: nil,
		}

		return state, nil
	}

	// apply changes to data
	newData := data.Data()
	maps.Copy(newData, req.Diff)

	revisionDate := req.UpdatedAt
	if revisionDate == nil {
		revisionDate = app.Pointer(req.CreatedAt)
	}

	state = &EntityChange{
		ID:           entityID,
		Data:         newData,
		Revision:     req.Revision,
		RevisionDate: revisionDate,
	}

	return state, nil
}

// UpdateEntityChangeRequest creates or updates a pending entity change request.
//
// If no pending change request exists for the given entity and user, a new one
// is created with revision 1. If a pending request already exists and the diff
// has changed, the existing request is updated and its revision is incremented.
// If the diff is identical, no action is performed.
func (s *Service) UpdateEntityChangeRequest(ctx context.Context, m *ds.EntityChangeRequest) (err error) {
	ctx, span := s.tracer.Start(ctx, "UpdateEntityChangeRequest")
	defer span.End()

	req, err := s.db.FindPendingEntityChangeRequest(ctx, m.EntityID, m.UserID)
	if err != nil {
		return err
	}

	if req == nil {
		m.Revision = 1
		return s.db.CreateEntityChangeRequest(ctx, m)
	}

	if !hasDiff(req.Diff, m.Diff) {
		return nil
	}

	m.ID = req.ID
	m.Revision = req.Revision + 1
	m.Status = req.Status
	m.CreatedAt = req.CreatedAt
	m.UpdatedAt = app.Pointer(time.Now())

	err = s.db.UpdateEntityChangeRequest(ctx, m)
	return
}
