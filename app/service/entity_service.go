package service

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/jackc/pgx/v5"
)

// DefinitionType
// Definitions is a special kind of data that displayed only if one topic selected
// Should filter out definitions from regular data
const DefinitionType = "definition"

// DefinitionPathPrefix
// To get needed definition I need to know its path
// So all definitions should be stored in one location
// and it must be persistent
// TODO make it configurable?
const DefinitionPathPrefix = "definitions/"

// SingleTopicOrder
// when only one topic selected sort data in this order
const SingleTopicOrder = "'person', 'book', 'conference', 'software'"

type FilterEntitiesParams struct {
}

func FilterEntities(req FilterEntitiesParams) (data []ds.Entity, count int, err error) {
	//q := database.ORM().
	//	Model(&data).
	//	Apply(filter.PageFilter(req.Page, req.Limit))
	//
	//if req.Name != "" {
	//	q.Where("title ILIKE ?", "%"+req.Name+"%")
	//}
	//if req.Addr != "" {
	//	q.Where("data ->> 'home_addr' IS NOT NULL AND data ->> 'home_addr' ILIKE ?", "%"+req.Addr+"%")
	//}
	//if req.Query != "" {
	//	// TODO this query will also match keys, but only values needed
	//	q.Where("data::text  ILIKE ?", "%"+req.Query+"%")
	//}
	//
	//if len(req.Topics) > 0 {
	//	q.Join("JOIN entity_topics et ON et.entity_id=entity.id").
	//		Join("JOIN topics t ON et.topic_id=t.id").
	//		Group("entity.id")
	//}
	//
	//if len(req.Topics) == 1 {
	//	q.Where("t.name = ?", req.Topics[0]).
	//		// Add specific order when only one topic is selected
	//		OrderExpr(fmt.Sprintf("array_position(array[%s]::text[], type)", SingleTopicOrder))
	//}
	//
	//if len(req.Topics) > 1 {
	//	q.WhereIn("t.name IN (?)", req.Topics).
	//		Having("COUNT(t.id) = ?", len(req.Topics))
	//}
	//
	//// should not match definitions
	//if len(req.Topics) > 0 {
	//	q.Where("type != ?", DefinitionType)
	//}
	//
	//if req.Collection != 0 {
	//	q.Join("JOIN collection_entities ce ON ce.entity_id=entity.id").
	//		Where("ce.collection_id = ?", req.Collection)
	//}
	//
	//// filter by exact repo if set
	//if req.Repo != 0 {
	//	q.Where("entity.repo_id = ?", req.Repo)
	//} else {
	//	// otherwise filter by user's repos && global
	//	// this will be not efficient once we'll have lots of global repos
	//	// add repo_type to entity maybe?
	//	if req.User == 0 {
	//		// filter only by global repos
	//		q.Where("entity.repo_id IN (SELECT id FROM repositories WHERE type = ?)", model.RepoTypeGlobal)
	//	} else {
	//		// filter by global repos and any that belongs to user
	//		q.Where(
	//			"entity.repo_id IN (SELECT id FROM repositories WHERE type = ? OR user_id = ?)",
	//			model.RepoTypeGlobal,
	//			req.User,
	//		)
	//	}
	//}
	//
	//if req.WithRepo {
	//	q.Apply(WithRepository())
	//}
	//
	//q.OrderExpr("updated_at DESC, created_at DESC")
	//count, err = q.SelectAndCount()

	return
}

// Definition returns definition of given name
func Definition(name string) (def *ds.Entity, err error) {
	//def = &model.Entity{}
	//err = database.ORM().
	//	Model(def).
	//	Where("type = ?", DefinitionType).
	//	Where("path = ?", DefinitionPathPrefix+name).
	//	First()
	//
	//if err == pg.ErrNoRows {
	//	def = nil
	//	err = nil
	//}

	return
}

func CreateOrUpdateEntity(ctx context.Context, elem *ds.Entity) (err error) {
	old := ds.Entity{}
	err = pgxscan.Get(ctx, app.DB(), &old, `SELECT * FROM entities WHERE path = $1`, elem.Path)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return
	}

	// create new
	if errors.Is(err, pgx.ErrNoRows) {
		r := app.DB().QueryRow(ctx,
			"INSERT INTO entities (path, title, type, data, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			elem.Path,
			elem.Title,
			elem.Type,
			elem.Data,
			time.Now(),
		)

		err = r.Scan(&elem.ID)
		if err != nil {
			return
		}

		return
	}

	// update old
	_, err = app.DB().Exec(ctx,
		// TODO updated_at should be set if entity is changed for sure
		"UPDATE entities SET title=$1, type=$2, data=$3, updated_at=NOW(), deleted_at=NULL WHERE id=$4",
		elem.Title,
		elem.Type,
		elem.Data,
		old.ID,
	)
	if err != nil {
		return
	}

	elem.ID = old.ID
	return nil
}

//func FindByID(id int64, filters ...filter.Fn) (m ds.Entity, err error) {
//q := database.ORM().
//	Model(&m).
//	Where("entity.id = ?", id)
//
//filter.Apply(q, filters...)
//
//err = q.First()
//return
//}

//func WithRepository() filter.Fn {
//	return func(q *orm.Query) (*orm.Query, error) {
//		q.Relation("Repo")
//		return q, nil
//	}
//}
