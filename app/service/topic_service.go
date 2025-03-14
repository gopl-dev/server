package service

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/jackc/pgx/v5"
)

var ErrNoRows = errors.New("no rows in result set")

type FilterTopicsParams struct{}

func FilterTopics(req FilterTopicsParams) (data []ds.Topic, count int, err error) {
	//q := database.ORM().
	//	Model(&data)
	//
	//if req.Name != "" {
	//	q.Where("name ILIKE ?", req.Name+"%")
	//}
	//
	//count, err = q.SelectAndCount()
	return
}

type CommonTopicsParams struct {
	Topics       []string
	CollectionID int64
	RepoID       int64
}

// CommonTopics return topics that in common with entities from given topics
// For example:
//
//	entity1 have topics A, B, C
//	entity2 have topics A, D, E
//
// For given topic A, should return B, C, D, E
// For given topic B, should return A, C
// For given topics [A, B] should return C
// For given topics [A, E] should return D
func CommonTopics(in CommonTopicsParams) (out []string, err error) {
	// just return all topics
	if len(in.Topics) == 0 && in.CollectionID == 0 && in.RepoID == 0 {
		//err = database.ORM().
		//	Model(&[]model.Topic{}).
		//	Column("name").
		//	Order("name").
		//	Select(&out)
		return
	}

	// just return all topics of collection
	if in.CollectionID != 0 && len(in.Topics) == 0 && in.RepoID == 0 {
		//err = database.ORM().
		//	Model(&[]model.Topic{}).
		//	Column("name").
		//	Join("JOIN entity_topics et ON et.topic_id=topic.id").
		//	Join("JOIN collection_entities ce ON ce.entity_id=et.entity_id").
		//	Where("ce.collection_id = ?", in.CollectionID).
		//	Order("name").
		//	Group("topic.id").
		//	Select(&out)
		return
	}

	// just return all topics of repo
	if in.RepoID != 0 && len(in.Topics) == 0 && in.CollectionID == 0 {
		//err = database.ORM().
		//	Model(&[]model.Topic{}).
		//	Where("repo_id = ?", in.RepoID).
		//	Column("name").
		//	Order("name").
		//	Select(&out)
		return
	}

	// select entities that having given topics
	//entities := database.ORM().
	//	Model(&[]model.Topic{}).
	//	Column("e.id").
	//	Join("JOIN entity_topics et ON et.topic_id=topic.id").
	//	Join("JOIN entities e ON et.entity_id=e.id").
	//	WhereIn("topic.name IN (?)", in.Topics).
	//	Having("COUNT(topic.id) = ?", len(in.Topics)).
	//	Group("e.id")
	//
	//if in.CollectionID != 0 {
	//	entities = entities.
	//		Join("JOIN collection_entities ce ON ce.entity_id=e.id").
	//		Where("ce.collection_id = ?", in.CollectionID)
	//}
	//if in.RepoID != 0 {
	//	entities = entities.
	//		Where("topic.repo_id = ?", in.RepoID)
	//}

	// select topics (of selected entities) that not a given topics
	//err = database.ORM().
	//	Model().
	//	Column("t.name").
	//	With("selected_entities", entities).
	//	TableExpr("selected_entities se").
	//	Join("JOIN entity_topics et ON et.entity_id=se.id").
	//	Join("JOIN topics t ON et.topic_id=t.id").
	//	WhereIn("t.name NOT IN (?)", in.Topics).
	//	Order("name").
	//	Group("t.id").
	//	Select(&out)

	return
}

func FirstOrCreateTopic(ctx context.Context, m *ds.Topic) (err error) {
	db := app.DB()
	err = pgxscan.Get(ctx, db, m, `SELECT * FROM topics WHERE name = $1 LIMIT 1`, m.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		r := app.DB().QueryRow(ctx, "INSERT INTO topics (name) VALUES ($1) RETURNING id", m.Name)
		err = r.Scan(&m.ID)
		if err != nil {
			return
		}
	}

	return
}

func CreateEntityTopic(ctx context.Context, m ds.EntityTopic) (err error) {
	_, err = app.DB().Exec(ctx,
		"INSERT INTO entity_topics (entity_id, topic_id) VALUES ($1, $2)",
		m.EntityID,
		m.TopicID,
	)

	return
}
