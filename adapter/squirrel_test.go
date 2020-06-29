package adapter

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/query"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testCase struct {
	query *query.Query
	sql   string
	args  []interface{}
}

var metaStub = &entity.MetaInfo{
	Fields: map[string]*entity.FieldInfo{
		"id": {
			DbAlias:     "id",
			FullDbAlias: "test_table.id",
		},
	},
	TableName: "test_table",
}

var testCases = []testCase{
	{
		query.NewQuery(metaStub).AndWhere("id", "=", 1).OrWhere("id", "=", 3).Limit(1),
		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE (id = $1 OR id = $2) LIMIT 1",
		[]interface{}{1, 3},
	},
	{
		query.NewQuery(metaStub).AndWhere("id", "=", 1).AndWhere("id", "=", 2),
		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE (id = $1 AND id = $2)",
		[]interface{}{1, 2},
	},
	{
		query.NewQuery(metaStub).AndWhere("id", "IN", 1, 2, 3, 4),
		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE id IN ($1,$2,$3,$4)",
		[]interface{}{1, 2, 3, 4},
	},
	{
		query.NewQuery(metaStub).AndWhere("id", "=", 1).OrWhere("id", "=", 3).Limit(1).
			Union(query.NewQuery(metaStub).AndWhere("id", "=", 5)),
		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE (id = $1 OR id = $2) LIMIT 1 UNION SELECT test_table.id as \"test_table.id\" FROM test_table WHERE id = $3",
		[]interface{}{1, 3, 5},
	},
	{
		query.NewQuery(metaStub).AndWhere("id", "=", 1).AndNestedWhere(func(q *query.Query) {
			q.AndWhere("id", ">", 2).AndWhere("id", "<", 10)
		}),
		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE (id = $1 AND (id > $2 AND id < $3))",
		[]interface{}{1, 2, 10},
	},
	{
		query.NewQuery(metaStub).AndWhere("id", "=", 1).
			AndNestedWhere(func(q *query.Query) {
				q.AndWhere("id", ">", 2).AndWhere("id", "<", 10)
			}).
			OrNestedWhere(func(q *query.Query) {
				q.AndWhere("id", ">", 2).OrNestedWhere(func(q *query.Query) {
					q.AndWhere("id", ">", 0).OrWhere("id", "<", 40)
				})
			}),

		"SELECT test_table.id as \"test_table.id\" FROM test_table WHERE ((id = $1 AND (id > $2 AND id < $3)) OR (id > $4 OR (id > $5 OR id < $6)))",
		[]interface{}{1, 2, 10, 2, 0, 40},
	},
}

func TestQueryToSquirrelSql(t *testing.T) {
	for _, tCase := range testCases {
		sql, args, _ := QueryToSql(tCase.query)
		assert.Equal(t, tCase.sql, sql)
		assert.Equal(t, tCase.args, args)
	}
}
