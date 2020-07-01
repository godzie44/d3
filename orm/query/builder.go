package query

import (
	"errors"
	"fmt"
	"github.com/godzie44/d3/orm/entity"
	"sort"
	"strings"
)

var (
	ErrRelatedEntityNotFound = errors.New("related entity not found")
	ErrRelatedFieldNotFound  = errors.New("related field not found")
)

type Union struct {
	Q *Query
}

type Where struct {
	Field  string
	Op     string
	Params []interface{}
}

type AndWhere struct {
	Where
}

type OrWhere struct {
	Where
}

type NestedWhere struct {
	Supply func(query *Query)
}

type AndNestedWhere struct {
	NestedWhere
}

type OrNestedWhere struct {
	NestedWhere
}

type Having struct {
	Field  string
	Op     string
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

	columns Columns
	from    From
	where   []interface{}
	having  []*Having
	join    []*Join
	union   []*Union
	group   GroupBy
	orderBy Order

	limit  Limit
	offset Offset
}

// ForEntity - create new query.
func New() *Query {
	return &Query{
		relationsMeta: make(map[entity.Name]*entity.MetaInfo),
		withList:      make(map[entity.Name]struct{}),
	}
}

// ForEntity - bind entity to query.
func (q *Query) ForEntity(targetEntityMeta *entity.MetaInfo) *Query {
	q.mainMeta = targetEntityMeta
	q.From(targetEntityMeta.TableName).
		addEntityFieldsToSelect(targetEntityMeta)
	return q
}

// From - set table name in FROM query section.
func (q *Query) From(tableName string) *Query {
	q.from = From(tableName)
	return q
}

// Select - add columns to SELECT query section.
func (q *Query) Select(columns ...string) *Query {
	q.columns = append(q.columns, columns...)
	return q
}

func (q *Query) ownerMeta() *entity.MetaInfo {
	return q.mainMeta
}

func (q *Query) addEntityFieldsToSelect(meta *entity.MetaInfo) {
	fields := make([]*entity.FieldInfo, 0, len(meta.Fields))
	for _, field := range meta.Fields {
		fields = append(fields, field)
	}

	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Name > fields[j].Name
	})

	for _, f := range fields {
		q.Select(f.FullDbAlias)
	}
	for _, rel := range meta.OneToOneRelations() {
		q.Select(meta.FullColumnAlias(rel.JoinColumn))
	}
}

// AndWhere add WHERE expression in select query.
// Example:
// q.Where("a", "=", 1) - generate sql: WHERE a=?
//
// q.Where("a", "IS NOT NULL") - generate sql: WHERE a IS NOT NULL
func (q *Query) Where(field, operator string, params ...interface{}) *Query {
	q.where = append(q.where, &AndWhere{Where{
		Field:  field,
		Op:     strings.TrimSpace(strings.ToUpper(operator)),
		Params: params,
	}})

	return q
}

// AndWhere join WHERE expression in select query with AND operator.
// Example:
// q.AndWhere("a", "=", 1).AndWhere("b", "=",2) - generate sql: WHERE a=? AND b=?
//
// q.AndWhere("a", "IS NOT NULL") - generate sql: WHERE a IS NOT NULL
func (q *Query) AndWhere(field, operator string, params ...interface{}) *Query {
	q.where = append(q.where, &AndWhere{Where{
		Field:  field,
		Op:     strings.TrimSpace(strings.ToUpper(operator)),
		Params: params,
	}})

	return q
}

// OrWhere join WHERE expression in select query with OR operator.
// Example:
// q.AndWhere("a", "=", 1).OrWhere("b", "=",2) - generate sql: WHERE a=? OR b=?
//
// q.OrWhere("a", "IS NOT NULL") - generate sql: WHERE a IS NOT NULL
func (q *Query) OrWhere(field, operator string, params ...interface{}) *Query {
	q.where = append(q.where, &OrWhere{Where{
		Field:  field,
		Op:     strings.TrimSpace(strings.ToUpper(operator)),
		Params: params,
	}})
	return q
}

// AndNestedWhere join nested WHERE expression in select query with AND operator.
// Example:
// q.AndWhere("a", "=", 1).AndNestedWhere(func(q *Query){
//     q.OrWhere("b", "=", 2).OrWhere("c", "=", 3)
// }) - generate sql: WHERE a=? AND (b=? OR c=?)
func (q *Query) AndNestedWhere(f func(q *Query)) *Query {
	q.where = append(q.where, &AndNestedWhere{NestedWhere{Supply: f}})
	return q
}

// OrNestedWhere join nested WHERE expression in select query with OR operator.
// Example:
// q.OrWhere("a", "=", 1).OrNestedWhere(func(q *Query){
//     q.AndWhere("b", "=", 2).AndWhere("c", "=", 3)
// }) - generate sql: WHERE a=? OR (b=? AND c=?)
func (q *Query) OrNestedWhere(f func(q *Query)) *Query {
	q.where = append(q.where, &OrNestedWhere{NestedWhere{Supply: f}})
	return q
}

// GroupBy - add GROUP BY clause to query.
func (q *Query) GroupBy(expr string) *Query {
	q.group = GroupBy(expr)
	return q
}

// Having - add HAVING clause to query.
func (q *Query) Having(field, operator string, params ...interface{}) *Query {
	q.having = append(q.having, &Having{
		Field:  field,
		Op:     operator,
		Params: params,
	})
	return q
}

// Join - add JOIN clause to query.
// Example:
// q.Join(JoinRight, "profile", "user.id=profile.user_id")
func (q *Query) Join(joinType JoinType, table string, on string) *Query {
	q.join = append(q.join, &Join{
		Join: table,
		On:   on,
		Type: joinType,
	})
	return q
}

// Union - add UNION operator to query.
// Example:
// q.Union(q2.AndWhere("a=?", 1))
func (q *Query) Union(query *Query) *Query {
	q.union = append(q.union, &Union{
		Q: query,
	})
	return q
}

// Limit - add LIMIT clause to query.
func (q *Query) Limit(l int) *Query {
	q.limit = Limit(l)
	return q
}

// Offset - add OFFSET clause to query.
func (q *Query) Offset(o int) *Query {
	q.offset = Offset(o)
	return q
}

// OrderBy - add ORDER BY clause to query.
func (q *Query) OrderBy(stmts ...string) *Query {
	q.orderBy = stmts
	return q
}

// With - d3 will load with main entity related entities in same query.
// Example:
// q.With("myPkg/Entity2")
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

func Visit(q *Query, visitor func(pred interface{})) {
	visitor(q.from)
	visitor(q.columns)
	visitor(q.orderBy)

	for _, where := range q.where {
		visitor(where)
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
