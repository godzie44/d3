package adapter

import (
	"context"
	"d3/orm/entity"
	"d3/orm/query"
	"database/sql"
	_ "github.com/lib/pq"
)

type GoPgAdapter struct {
	pgDb         *sql.DB
	queryAdapter *SquirrelAdapter

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
		pgDb:         pgDb,
		queryAdapter: queryAdapter,
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

func (g *GoPgAdapter) Insert([]interface{}, *entity.MetaInfo) {
	//g.actionRegistry[persistence.ActionInsert] = action
}

func (g *GoPgAdapter) Update(entity []interface{}, meta *entity.MetaInfo) {
	//g.actionRegistry[persistence.ActionUpdate] = entity
}

func (g *GoPgAdapter) Remove(entity []interface{}, meta *entity.MetaInfo) {
	//g.actionRegistry[persistence.ActionDelete] = entity
}

func (g *GoPgAdapter) DoInTransaction(func() error) error {
	tx, err := g.pgDb.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	tx.Exec("")
	return tx.Commit()
	//txFunc := func(tx *pg.Tx) error {
	//	f()
	//
	//	for ActionType, arg := range g.actionRegistry {
	//		var err error
	//		switch ActionType {
	//		case ActionInsert:
	//			err = tx.Insert(arg)
	//		case ActionUpdate:
	//			err = tx.Update(arg)
	//		case ActionDelete:
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
