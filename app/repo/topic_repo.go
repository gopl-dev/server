package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrTopicFound is a sentinel error returned when topic not found.
	ErrTopicFound = app.ErrNotFound("topic not found")
)

// FilterTopics retrieves a paginated list of topics matching the given filter.
func (r *Repo) FilterTopics(ctx context.Context, f ds.TopicsFilter) (data []ds.Topic, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterTopic")
	defer span.End()

	if f.OrderBy == "" {
		f.OrderBy = "name"
		f.OrderDirection = "asc"
	}

	count, err = r.filter("topics").
		columns(`*`).
		where("type", f.Type).
		paginate(f.Page, f.PerPage).
		order(f.OrderBy, f.OrderDirection).
		withCount(f.WithCount).
		apply(whereIn("public_id", f.PublicIDs)).
		scan(ctx, &data)

	return
}

// CreateEntityTopic creates a single association between an entity and a topic.
func (r *Repo) CreateEntityTopic(ctx context.Context, entityID, topicID ds.ID) error {
	_, span := r.tracer.Start(ctx, "CreateEntityTopic")
	defer span.End()

	return r.insert(ctx, "entity_topics", data{
		"entity_id": entityID,
		"topic_id":  topicID,
	})
}

// CreateTopic inserts a new topic record into the database.
func (r *Repo) CreateTopic(ctx context.Context, t *ds.Topic) error {
	_, span := r.tracer.Start(ctx, "CreateTopic")
	defer span.End()

	return r.insert(ctx, "topics", data{
		"id":          t.ID,
		"type":        t.Type,
		"public_id":   t.PublicID,
		"name":        t.Name,
		"description": t.Description,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
		"deleted_at":  t.DeletedAt,
	})
}

// AttachTopics creates associations between an entity and the given topics.
func (r *Repo) AttachTopics(ctx context.Context, entityID ds.ID, topics []ds.Topic) error {
	_, span := r.tracer.Start(ctx, "AttachTopics")
	defer span.End()

	if len(topics) == 0 {
		return nil
	}

	ts := make([]data, len(topics))
	for i, t := range topics {
		ts[i] = data{
			"entity_id": entityID,
			"topic_id":  t.ID,
		}
	}

	return r.insert(ctx, "entity_topics", ts...)
}

// DetachTopics drops associations between an entity and topics.
func (r *Repo) DetachTopics(ctx context.Context, entityID ds.ID) error {
	_, span := r.tracer.Start(ctx, "DetachTopics")
	defer span.End()

	const q = "DELETE FROM entity_topics WHERE entity_id = $1"
	return r.exec(ctx, q, entityID)
}

// EntityTopics returns all non-deleted topics associated with the given entity.
func (r *Repo) EntityTopics(ctx context.Context, entityID ds.ID) ([]ds.Topic, error) {
	_, span := r.tracer.Start(ctx, "EntityTopics")
	defer span.End()

	topics := make([]ds.Topic, 0)
	const query = `SELECT t.* FROM entity_topics e JOIN topics t ON t.id = e.topic_id WHERE e.entity_id = $1 AND t.deleted_at IS NULL`

	err := pgxscan.Select(ctx, r.getDB(ctx), &topics, query, entityID)
	return topics, err
}
