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
	ID        ds.ID          `json:"id"`
	Data      map[string]any `json:"data"`
	Revision  int            `json:"revision"`
	UpdatedAt *time.Time     `json:"updated_at"`
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
			ID:        entityID,
			Data:      data.Data(),
			Revision:  0,
			UpdatedAt: nil,
		}

		return state, nil
	}

	// apply changes to data
	newData := data.Data()
	maps.Copy(newData, req.Diff)

	state = &EntityChange{
		ID:        entityID,
		Data:      newData,
		Revision:  req.Revision,
		UpdatedAt: req.UpdatedAt,
	}

	return state, nil
}
