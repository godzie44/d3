package query

import (
	"errors"
	"fmt"
	"github.com/godzie44/d3/orm/entity"
)

var (
	ErrRelatedEntityNotFound = errors.New("related entity not found")
	ErrRelatedFieldNotFound  = errors.New("related field not found")
)

type Union struct {
	UType string
	Q     *Query
}

type Where struct {
	Expr   string
	Params []interface{}
}

type AndWhere struct {
	Where
}

type OrWhere struct {
	Where
}

type Having struct {
	Expr   string
	Params []interface{}
}

type JoinType int

const (
	_ JoinType = iota
	JoinLeft
	JoinRight
	JoinInner
)

type Join struct {
	Join string
	On   string
	Type JoinType
}

type GroupBy string
type From string
type Limit int
type Offset int

type Columns []string
type Order []string

type Query struct {
	mainMeta      *entity.MetaInfo
	relationsMeta map[entity.Name]*entity.MetaInfo
	withList      map[entity.Name]struct{}

	columns  Columns
	from     From
	andWhere []*AndWhere
	orWhere  []*OrWhere
	having   []*Having
	join     []*Join
	union    []*Union
	group    GroupBy
	orderBy  Order

	limit  Limit
	offset Offset
}

func NewQuery(targetEntityMeta *entity.MetaInfo) *Query {
	q := &Query{
		relationsMeta: make(map[entity.Name]*entity.MetaInfo),
		withList:      make(map[entity.Name]struct{}),
	}

	return q.forEntity(targetEntityMeta)
}

func (q *Query) forEntity(meta *entity.MetaInfo) *Query {
	q.mainMeta = meta
	q.from = From(q.mainMeta.TableName)
	q.addEntityFieldsToSelect(meta)
	return q
}

func (q *Query) OwnerMeta() *entity.MetaInfo {
	return q.mainMeta
}

func (q *Query) addEntityFieldsToSelect(meta *entity.MetaInfo) {
	for _, field := range meta.Fields {
		q.columns = append(q.columns, field.FullDbAlias)
	}

	for _, rel := range meta.OneToOneRelations() {
		q.columns = append(q.columns, meta.FullColumnAlias(rel.JoinColumn))
	}
}

func (q *Query) AndWhere(expr string, params ...interface{}) *Query {
	q.andWhere = append(q.andWhere, &AndWhere{Where{
		Expr:   expr,
		Params: params,
	}})

	return q
}

func (q *Query) OrWhere(expr string, params ...interface{}) *Query {
	q.orWhere = append(q.orWhere, &OrWhere{Where{
		Expr:   expr,
		Params: params,
	}})
	return q
}

func (q *Query) GroupBy(expr string) *Query {
	q.group = GroupBy(expr)
	return q
}

func (q *Query) Having(expr string, params ...interface{}) *Query {
	q.having = append(q.having, &Having{
		Expr:   expr,
		Params: params,
	})
	return q
}

func (q *Query) Join(joinType JoinType, table string, on string) *Query {
	q.join = append(q.join, &Join{
		Join: table,
		On:   on,
		Type: joinType,
	})
	return q
}

func (q *Query) Union(query *Query, utype string) *Query {
	q.union = append(q.union, &Union{
		Q:     query,
		UType: utype,
	})
	return q
}

func (q *Query) Limit(l int) *Query {
	q.limit = Limit(l)
	return q
}

func (q *Query) Offset(o int) *Query {
	q.offset = Offset(o)
	return q
}

func (q *Query) OrderBy(stmts ...string) *Query {
	q.orderBy = stmts
	return q
}

func (q *Query) With(entityName entity.Name) error {
	defer func() {
		q.withList[entityName] = struct{}{}
	}()

	if _, exists := q.mainMeta.RelatedMeta[entityName]; exists {
		return q.joinEntity(entityName, q.mainMeta)
	}

	for _, meta := range q.relationsMeta {
		if _, exists := meta.RelatedMeta[entityName]; exists {
			return q.joinEntity(entityName, meta)
		}
	}

	return fmt.Errorf("%w: %s", ErrRelatedEntityNotFound, entityName)
}

func (q *Query) Visit(visitor func(pred interface{})) {
	visitor(q.from)
	visitor(q.columns)
	visitor(q.orderBy)

	for _, andWhere := range q.andWhere {
		visitor(andWhere)
	}
	for _, orWhere := range q.orWhere {
		visitor(orWhere)
	}
	for _, having := range q.having {
		visitor(having)
	}
	for _, join := range q.join {
		visitor(join)
	}
	for _, union := range q.union {
		visitor(union)
	}
	if string(q.group) != "" {
		visitor(q.group)
	}
	if int(q.limit) != 0 {
		visitor(q.limit)
	}
	if int(q.offset) != 0 {
		visitor(q.offset)
	}
}

func (q *Query) joinEntity(name entity.Name, ownerMeta *entity.MetaInfo) error {
	relatedEntityMeta, exists := ownerMeta.RelatedMeta[name]
	if !exists {
		return fmt.Errorf("%s: %w", name, ErrRelatedEntityNotFound)
	}

	var relation entity.Relation
	{
		for _, relation = range ownerMeta.Relations {
			if relation.RelatedWith().Equal(name) {
				break
			}
		}
	}

	if relation == nil {
		return fmt.Errorf("%s: %w", name, ErrRelatedFieldNotFound)
	}

	switch rel := relation.(type) {
	case *entity.OneToOne:
		q.Join(JoinLeft, relatedEntityMeta.TableName, fmt.Sprintf(
			"%s = %s",
			ownerMeta.FullColumnAlias(rel.JoinColumn), relatedEntityMeta.FullColumnAlias(rel.ReferenceColumn),
		))

	case *entity.OneToMany:
		q.Join(JoinLeft, relatedEntityMeta.TableName, fmt.Sprintf(
			"%s = %s",
			ownerMeta.Pk.FullDbAlias(), relatedEntityMeta.FullColumnAlias(rel.JoinColumn),
		))

	case *entity.ManyToMany:
		q.
			Join(JoinLeft, rel.JoinTable, fmt.Sprintf(
				"%s = %s.%s",
				ownerMeta.Pk.FullDbAlias(), rel.JoinTable, rel.JoinColumn,
			)).
			Join(JoinLeft, relatedEntityMeta.TableName, fmt.Sprintf(
				"%s.%s = %s",
				rel.JoinTable, rel.ReferenceColumn, relatedEntityMeta.Pk.FullDbAlias(),
			))
	}

	q.addEntityFieldsToSelect(relatedEntityMeta)
	q.relationsMeta[name] = relatedEntityMeta

	return nil
}
