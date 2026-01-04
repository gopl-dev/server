package repo

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// CreateEntityChangeLog inserts a new audit record.
func (r *Repo) CreateEntityChangeLog(ctx context.Context, log *ds.EntityChangeLog) error {
	const sql = `
		INSERT INTO entity_change_logs (id, entity_id, user_id, action, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	err := r.exec(ctx, sql, log.ID, log.EntityID, log.UserID, log.Action, log.CreatedAt)
	return err
}
