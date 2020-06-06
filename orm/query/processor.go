package query

import (
	"github.com/godzie44/d3/orm/entity"
	"strings"
)

var Preprocessor preprocessor

type preprocessor struct{}

func (preprocessor) CreateFetchPlan(q *Query) *FetchPlan {
	return &FetchPlan{
		query:         q,
		pks:           extractIdsIfPossible(q),
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
				if fields[i] == q.OwnerMeta().Pk.FullDbAlias() {
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

func (e *FetchPlan) Query() *Query {
	return e.query
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
