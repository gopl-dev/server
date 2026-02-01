package service

import (
	"context"

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
