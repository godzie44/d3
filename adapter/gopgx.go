package adapter

import (
	"context"
	"d3/orm/query"
	"github.com/jackc/pgx/v4"
)

type GoPgXAdapter struct {
	pgDb           *pgx.Conn
	actionRegistry map[int]interface{}
	queryAdapter   *SquirrelAdapter

	beforeQCallback, afterQCallback []func(query string, args ...interface{})
}

func NewGoPgXAdapter(pgDb *pgx.Conn, queryAdapter *SquirrelAdapter) *GoPgXAdapter {
	//dbConn, err := pgx.Connect(context.Background(), "")
	//if err != nil {
	//	panic(err)
	//}

	return &GoPgXAdapter{
		pgDb:           pgDb,
		actionRegistry: make(map[int]interface{}),
		queryAdapter:   queryAdapter,
	}
}

func (g *GoPgXAdapter) BeforeQuery(fn func(query string, args ...interface{})) {
	g.beforeQCallback = append(g.beforeQCallback, fn)
}

func (g *GoPgXAdapter) AfterQuery(fn func(query string, args ...interface{})) {
	g.afterQCallback = append(g.afterQCallback, fn)
}

func (g *GoPgXAdapter) ExecuteQuery(query *query.Query) ([]map[string]interface{}, error) {
	q, args, err := g.queryAdapter.ToSql(query)
	if err != nil {
		return nil, err
	}

	for i := range g.beforeQCallback {
		g.beforeQCallback[i](q, args...)
	}

	rows, err := g.pgDb.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}

	for i := range g.afterQCallback {
		g.afterQCallback[i](q, args...)
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range rows.FieldDescriptions() {
			m[string(col.Name)] = values[i]
		}

		result = append(result, m)
	}

	return result, nil
}

func (g *GoPgXAdapter) Insert(entity interface{}) error {
	g.actionRegistry[actionInsert] = entity
	return nil
}

func (g *GoPgXAdapter) Update(entity interface{}) error {
	g.actionRegistry[actionUpdate] = entity
	return nil
}

func (g *GoPgXAdapter) Remove(entity interface{}) error {
	g.actionRegistry[actionDelete] = entity
	return nil
}

func (g GoPgXAdapter) DoInTransaction(func()) error {
	tx, err := g.pgDb.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	tx.Exec(context.Background(), "")
	return tx.Commit(context.Background())
}
