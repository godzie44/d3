package query

import (
	"github.com/godzie44/d3/orm/entity"
)

var Preprocessor preprocessor

type preprocessor struct{}

func (preprocessor) MakeFetchPlan(q *Query) *FetchPlan {
	return &FetchPlan{
		query:         q,
		pks:           extractIdsIfPossible(q),
		fetchWithList: getFetchList(q.mainMeta, q),
	}
}

func extractIdsIfPossible(q *Query) []interface{} {
	var idList []interface{}

	Visit(q, func(pred interface{}) {
		switch where := pred.(type) {
		case *OrWhere:
			if canExtractIdsFromWhere(where.Where, q.ownerMeta()) {
				idList = append(idList, where.Params...)
			}
		case *AndWhere:
			if canExtractIdsFromWhere(where.Where, q.ownerMeta()) {
				idList = append(idList, where.Params...)
			}
		default:
			return
		}
	})

	return idList
}

func canExtractIdsFromWhere(w Where, meta *entity.MetaInfo) bool {
	return (w.Field == meta.Pk.FullDbAlias() || w.Field == meta.Pk.Field.DbAlias) &&
		(w.Op == "=" || w.Op == "IN")
}

func getFetchList(meta *entity.MetaInfo, q *Query) []*executeWith {
	var result []*executeWith

	for _, rel := range meta.Relations {
		if _, exists := q.withList[rel.RelatedWith()]; !exists {
			continue
		}

		result = append(result, &executeWith{
			entityMeta: meta.RelatedMeta[rel.RelatedWith()],
			relation:   rel,
			withList:   getFetchList(meta.RelatedMeta[rel.RelatedWith()], q),
		})
	}

	return result
}

type FetchPlan struct {
	query         *Query
	pks           []interface{}
	fetchWithList []*executeWith
}

func (e *FetchPlan) NoNestedWhere() bool {
	return len(e.query.andNestedWhere) == 0 && len(e.query.orNestedWhere) == 0
}

func (e *FetchPlan) WhereExprCount() int {
	return len(e.query.andWhere) + len(e.query.orWhere)
}

func (e *FetchPlan) EntityName() entity.Name {
	if e.query.ownerMeta() != nil {
		return e.query.ownerMeta().EntityName
	}
	return ""
}

func (e *FetchPlan) PKs() []interface{} {
	return e.pks
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
