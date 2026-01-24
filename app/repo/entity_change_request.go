package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// FindPendingEntityChangeRequest ...
func (r *Repo) FindPendingEntityChangeRequest(ctx context.Context, entityID, userID ds.ID) (*ds.EntityChangeRequest, error) {
	ctx, span := r.tracer.Start(ctx, "FindPendingEntityChangeRequest")
	defer span.End()

	const query = `SELECT * FROM entity_change_requests WHERE entity_id = $1 AND user_id = $2 AND status = $3 ORDER BY updated_at DESC NULLS LAST, created_at DESC LIMIT 1`

	req := new(ds.EntityChangeRequest)
	err := pgxscan.Get(ctx, r.getDB(ctx), req, query, entityID, userID, ds.EntityChangePending)
	if noRows(err) {
		return nil, nil
	}

	return req, err
}
