package query

import (
	"d3/orm/entity"
	"strings"
)

var Preprocessor preprocessor

type preprocessor struct{}

func (preprocessor) Process(q *Query) error {
	meta := q.mainMeta

	err := handleEagerRelations(q, meta)
	if err != nil {
		return err
	}

	return nil
}

func handleEagerRelations(q *Query, meta *entity.MetaInfo) error {
	for _, field := range meta.Fields {
		if field.Relation != nil && field.Relation.IsEager() {
			relatedEntityName := field.Relation.RelatedWith()

			if q.inJoinedEntities(relatedEntityName) {
				continue
			}

			err := q.With(relatedEntityName)
			if err != nil {
				return err
			}
			err = handleEagerRelations(q, meta.RelatedMeta[relatedEntityName])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (preprocessor) CreateFetchPlan(q *Query) *FetchPlan {
	return &FetchPlan{
		query: q,
		entityIds: extractIdsIfPossible(q),
		fetchWithList: getFetchList(q.mainMeta, q),
	}
}

func extractIdsIfPossible(q *Query) []interface{} {
	isIdQuery := true
	var idList []interface{}

	q.Visit(func(pred interface{}) {
		switch where := pred.(type) {
		case *OrWhere:
		case *AndWhere:
			fields := strings.Fields(where.Expr)
			for i := range fields {
				if fields[i] == q.OwnerMeta().FullFieldAlias(q.OwnerMeta().PkField()) {
					idList = append(idList, where.Params...)
					return
				}
			}
			isIdQuery = false
		default:
			return
		}
	})

	if !isIdQuery {
		idList = []interface{}{}
	}

	return idList
}

func getFetchList(meta *entity.MetaInfo, q *Query) []*executeWith {
	var result []*executeWith

	for _, f := range meta.Fields {
		if !f.IsRelation() {
			continue
		}

		if _, exists := q.withList[f.Relation.RelatedWith()]; !exists {
			continue
		}

		result = append(result, &executeWith{
			entityMeta: meta.RelatedMeta[f.Relation.RelatedWith()],
			relation:   f.Relation,
			withList:   getFetchList(meta.RelatedMeta[f.Relation.RelatedWith()], q),
		})
	}

	return result
}

type FetchPlan struct {
	query         *Query
	entityIds     []interface{}
	fetchWithList []*executeWith
}

func (e *FetchPlan) Query() *Query {
	return e.query
}

func (e *FetchPlan) EntityIds() []interface{} {
	return e.entityIds
}

type executeWith struct {
	entityMeta *entity.MetaInfo
	relation   entity.Relation
	withList   []*executeWith
}

func (e *FetchPlan) HasJoins() bool {
	return len(e.fetchWithList) != 0
}

func (e *FetchPlan) CanFetchRelation(rel entity.Relation) bool {
	for _, with := range e.fetchWithList {
		if rel == with.relation {
			return true
		}
	}

	return false
}

func (e *FetchPlan) GetChildPlan(rel entity.Relation) *FetchPlan {
	for _, with := range e.fetchWithList {
		if rel == with.relation {
			return &FetchPlan{fetchWithList: with.withList}
		}
	}

	return &FetchPlan{}
}
