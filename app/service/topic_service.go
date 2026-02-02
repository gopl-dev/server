package service

import (
	"context"
	"fmt"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// FilterTopics retrieves a paginated list of topics matching the given filter.
func (s *Service) FilterTopics(ctx context.Context, f ds.TopicsFilter) (data []ds.Topic, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterTopics")
	defer span.End()

	return s.db.FilterTopics(ctx, f)
}

// AttachTopics associates the given topics with the specified entity.
func (s *Service) AttachTopics(ctx context.Context, entityID ds.ID, topics []ds.Topic) (err error) {
	ctx, span := s.tracer.Start(ctx, "AttachTopics")
	defer span.End()

	return s.db.AttachTopics(ctx, entityID, topics)
}

// normalizeTopics filters, deduplicates, and validates input topics
// against the allowed topics for the given entity type.
//
// Only topics that exist and belong to the specified entity type are kept;
// all others are silently discarded. Duplicate topic IDs are removed while
// preserving the original order.
func (s *Service) normalizeTopics(ctx context.Context, input []ds.Topic, typ ds.EntityType, minRequired int) (resolved []ds.Topic, err error) {
	topics, _, err := s.FilterTopics(ctx, ds.TopicsFilter{
		Type:    typ,
		PerPage: ds.PerPageNoLimit,
	})
	if err != nil {
		return nil, err
	}

	allowed := make(map[ds.ID]ds.Topic, len(topics))
	for _, t := range topics {
		allowed[t.ID] = t
	}

	seen := make(map[ds.ID]struct{}, len(input))
	resolved = make([]ds.Topic, 0, len(input))

	for _, t := range input {
		if _, dup := seen[t.ID]; dup {
			continue
		}
		seen[t.ID] = struct{}{}

		if at, ok := allowed[t.ID]; ok {
			resolved = append(resolved, at)
		}
	}

	if minRequired > 0 && len(resolved) < minRequired {
		return nil, app.NewInputError(
			"topics",
			fmt.Sprintf("at least %d topic(s) required", minRequired),
		)
	}

	return resolved, nil
}
