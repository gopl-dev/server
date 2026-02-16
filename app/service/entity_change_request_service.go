package service

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/ds/prop"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/diff"
	"github.com/gopl-dev/server/email"
)

var (
	// ErrChangeRequestAlreadyCommited indicates that a change request has already been applied and cannot be modified.
	ErrChangeRequestAlreadyCommited = errors.New("change request already committed")
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

	req, err := s.db.GetPendingChangeRequest(ctx, entityID, user.ID)
	if errors.Is(err, repo.ErrEntityChangeRequestNotFound) {
		// no changes were made yet, return data as is
		state = &EntityChange{
			ID:           entityID,
			Data:         data.Data(),
			Revision:     0,
			RevisionDate: nil,
		}

		return state, nil
	}
	if err != nil {
		return nil, err
	}

	// apply changes to data
	newData := data.Data()
	for k, v := range newData {
		if data.PropertyType(k).Patchable() {
			req.Diff[k], err = app.ApplyPatch(app.String(v), app.String(req.Diff[k]))
			if err != nil {
				return nil, err
			}
		}
	}
	maps.Copy(newData, req.Diff)

	revisionDate := req.UpdatedAt
	if revisionDate == nil {
		revisionDate = new(req.CreatedAt)
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
	ctx, span := s.tracer.Start(ctx, "UpdateChangeRequest")
	defer span.End()

	req, err := s.db.GetPendingChangeRequest(ctx, m.EntityID, m.UserID)
	if errors.Is(err, repo.ErrEntityChangeRequestNotFound) {
		m.Revision = 1
		return s.db.CreateChangeRequest(ctx, m)
	}
	if err != nil {
		return err
	}

	if !hasDiff(req.Diff, m.Diff) {
		return nil
	}

	m.ID = req.ID
	m.Revision = req.Revision + 1
	m.Status = req.Status
	m.CreatedAt = req.CreatedAt
	m.UpdatedAt = new(time.Now())

	err = s.db.UpdateChangeRequest(ctx, m)
	return
}

// FilterChangeRequests retrieves a paginated list of change requests matching the given filter.
func (s *Service) FilterChangeRequests(ctx context.Context, f ds.ChangeRequestsFilter) (data []ds.EntityChangeRequest, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterChangeRequests")
	defer span.End()

	return s.db.FilterChangeRequests(ctx, f)
}

// ChangeDiff represents the difference between current and proposed values in an entity change request.
// For type "diff" Diff property should be set, for other types Current and Proposed should be set.
type ChangeDiff struct {
	Key      string    `json:"key"`
	Type     prop.Type `json:"type"`
	Diff     string    `json:"diff,omitempty"`
	Current  any       `json:"current,omitempty"`
	Proposed any       `json:"proposed,omitempty"`
}

// GetChangeRequestDiff retrieves a change request and computes the diff between proposed and current values.
// Returns the diff containing both proposed changes and current values, along with the change request.
func (s *Service) GetChangeRequestDiff(ctx context.Context, reqID ds.ID) (diffs []ChangeDiff, req *ds.EntityChangeRequest, err error) {
	ctx, span := s.tracer.Start(ctx, "GetChangeRequestDiff")
	defer span.End()

	req, err = s.db.GetChangeRequestByID(ctx, reqID)
	if err != nil {
		return
	}

	entity, err := s.GetDataProviderFromEntityType(ctx, req.EntityID, req.EntityType)
	if err != nil {
		return
	}

	diffs, err = makeChangesDiff(entity, req.Diff)
	if err != nil {
		return
	}

	return diffs, req, nil
}

// ApplyChangeRequest applies a pending change request to its associated entity.
func (s *Service) ApplyChangeRequest(ctx context.Context, reqID ds.ID) (err error) {
	ctx, span := s.tracer.Start(ctx, "ApplyChangeRequest")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return app.ErrUnauthorized()
	}

	changes, req, err := s.GetChangeRequestDiff(ctx, reqID)
	if err != nil {
		return err
	}
	req.ReviewerID = new(user.ID)

	if req.Status == ds.EntityChangeCommitted {
		return ErrChangeRequestAlreadyCommited
	}

	switch req.EntityType {
	case ds.EntityTypeBook:
		err = s.ApplyChangesToBook(ctx, changes, req, true)
	case ds.EntityTypePage:
		err = s.ApplyChangesToPage(ctx, changes, req, true)
	default:
		err = ErrInvalidEntityType
	}

	return
}

// RejectChangeRequest rejects a pending change request with a review note.
func (s *Service) RejectChangeRequest(ctx context.Context, id, reviewerID ds.ID, note string) (err error) {
	ctx, span := s.tracer.Start(ctx, "RejectChangeRequest")
	defer span.End()

	req, err := s.db.GetChangeRequestByID(ctx, id)
	if err != nil {
		return
	}

	author, err := s.GetUserByID(ctx, req.UserID)
	if err != nil {
		return
	}

	entity, err := s.GetEntityByID(ctx, req.EntityID)
	if err != nil {
		return
	}

	err = s.db.RejectChangeRequest(ctx, id, reviewerID, note)
	if err != nil {
		return nil
	}

	return email.Send(author.Email, email.ChangesRejected{
		Username:    author.Username,
		EntityTitle: entity.Title,
		Note:        note,
		ViewURL:     entity.ViewURL(),
	})
}

// CommitChangeRequest marks a change request as committed in the database.
func (s *Service) CommitChangeRequest(ctx context.Context, req *ds.EntityChangeRequest) error {
	ctx, span := s.tracer.Start(ctx, "RejectChangeRequest")
	defer span.End()

	req.Status = ds.EntityChangeCommitted
	return s.db.CommitChangeRequest(ctx, req)
}

// GetDataProviderFromEntityType retrieves an entity by ID and type, returning it as a DataProvider interface.
func (s *Service) GetDataProviderFromEntityType(ctx context.Context, id ds.ID, t ds.EntityType) (dp ds.DataProvider, err error) {
	ctx, span := s.tracer.Start(ctx, "GetDataProviderFromEntityType")
	defer span.End()

	switch t {
	case ds.EntityTypeBook:
		dp, err = s.GetBookByID(ctx, id)
	case ds.EntityTypePage:
		dp, err = s.GetPageByID(ctx, id)
	default:
		err = ErrInvalidEntityType
	}

	return dp, err
}

func makeChangesDiff(orig ds.DataProvider, changes map[string]any) (diffs []ChangeDiff, err error) {
	data := orig.Data()
	diffs = make([]ChangeDiff, 0, len(data))
	for k := range data {
		v, ok := changes[k]
		if ok {
			d := ChangeDiff{
				Key:      k,
				Type:     orig.PropertyType(k),
				Diff:     "",
				Current:  nil,
				Proposed: nil,
			}

			if d.Type.Patchable() {
				dc, err := diff.ComputeFromPatch(app.String(data[k]), app.String(v))
				if err != nil {
					return nil, err
				}
				d.Diff = dc.HTML()
			} else {
				d.Current = data[k]
				d.Proposed = v
			}

			diffs = append(diffs, d)
		}
	}

	return diffs, nil
}
