package adapter

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"d3/orm/query"
)

type GoPgAdapter struct {
	pgDb           *sql.DB
	actionRegistry map[int]interface{}
	queryAdapter   *SquirrelAdapter

	beforeQCallback, afterQCallback func(query string, args ...interface{})
}

func (g *GoPgAdapter) BeforeQuery(fn func(query string, args ...interface{})) {
	g.beforeQCallback = fn
}

func (g *GoPgAdapter) AfterQuery(fn func(query string, args ...interface{})) {
	g.afterQCallback = fn
}

func NewGoPgAdapter(pgDb *sql.DB, queryAdapter *SquirrelAdapter) *GoPgAdapter {
	return &GoPgAdapter{
		pgDb:           pgDb,
		actionRegistry: make(map[int]interface{}),
		queryAdapter:   queryAdapter,
	}
}

func (g *GoPgAdapter) ExecuteQuery(query *query.Query) ([]map[string]interface{}, error) {
	q, args, err := g.queryAdapter.ToSql(query)
	if err != nil {
		return nil, err
	}

	if g.beforeQCallback != nil {
		g.beforeQCallback(q, args)
	}

	rows, err := g.pgDb.Query(q, args...)
	if err != nil {
		return nil, err
	}

	if g.afterQCallback != nil {
		g.afterQCallback(q, args)
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		result = append(result, m)
	}

	return result, nil
}

const (
	actionInsert = iota
	actionUpdate
	actionDelete
)

func (g *GoPgAdapter) Insert(entity interface{}) error {
	g.actionRegistry[actionInsert] = entity
	return nil
}

func (g *GoPgAdapter) Update(entity interface{}) error {
	g.actionRegistry[actionUpdate] = entity
	return nil
}

func (g *GoPgAdapter) Remove(entity interface{}) error {
	g.actionRegistry[actionDelete] = entity
	return nil
}

func (g *GoPgAdapter) DoInTransaction(f func()) error {
	tx, err := g.pgDb.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	tx.Exec("")
	return tx.Commit()
	//txFunc := func(tx *pg.Tx) error {
	//	f()
	//
	//	for action, arg := range g.actionRegistry {
	//		var err error
	//		switch action {
	//		case actionInsert:
	//			err = tx.Insert(arg)
	//		case actionUpdate:
	//			err = tx.Update(arg)
	//		case actionDelete:
	//			err = tx.Delete(arg)
	//		}
	//
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	g.actionRegistry = make(map[int]interface{})
	//
	//	return nil
	//}
	//return g.db.RunInTransaction(txFunc)
}