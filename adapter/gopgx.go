package adapter

import (
	"context"
	"d3/orm/persistence"
	"d3/orm/query"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strconv"
	"strings"
)

type GoPgXAdapter struct {
	pgDb         *pgx.Conn
	queryAdapter *SquirrelAdapter

	beforeQCallback, afterQCallback []func(query string, args ...interface{})
}

func NewGoPgXAdapter(pgDb *pgx.Conn, queryAdapter *SquirrelAdapter) *GoPgXAdapter {
	return &GoPgXAdapter{
		pgDb:         pgDb,
		queryAdapter: queryAdapter,
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

func (g *GoPgXAdapter) Insert(table string, cols []string, values []interface{}) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	_, err := g.pgDb.Exec(
		context.Background(),
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ",")),
		values...,
	)
	fmt.Println(
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ",")),
		values,
	)

	if err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}

	return nil
}

func (g *GoPgXAdapter) InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner persistence.Scanner) error) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	row := g.pgDb.QueryRow(
		context.Background(),
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ","), strings.Join(returnCols, ",")),
		values...,
	)
	fmt.Println(
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ","), strings.Join(returnCols, ",")),
		values,
	)

	if err := withReturned(row); err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}

	return nil
}

func (g *GoPgXAdapter) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
	queryValues := values

	setCommands := make([]string, len(values))
	var placeholderNum int
	for placeholderNum = 0; placeholderNum < len(values); placeholderNum++ {
		setCommands[placeholderNum] = cols[placeholderNum] + "=$" + strconv.Itoa(placeholderNum+1)
	}

	var whereStr string
	for col, val := range identityCond {
		queryValues = append(queryValues, val)
		whereStr += " " + col + "=$" + strconv.Itoa(placeholderNum+1)
		placeholderNum++
	}

	_, err := g.pgDb.Exec(
		context.Background(),
		fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setCommands, ","), whereStr),
		queryValues...,
	)

	fmt.Println(
		fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setCommands, ","), whereStr),
		queryValues,
	)

	if err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}
	return nil
}

func (g *GoPgXAdapter) Remove(table string, identityCond map[string]interface{}) error {
	args := make([]interface{}, 0, len(identityCond))
	where := make([]string, 0, len(identityCond))

	for col, val := range identityCond {
		args = append(args, val)
		where = append(where, col+"=$"+strconv.Itoa(len(args)))
	}

	_, err := g.pgDb.Exec(
		context.Background(),
		fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(where, " AND ")),
		args...,
	)

	fmt.Println(
		fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(where, " AND ")),
		args,
	)

	return err
}

func (g GoPgXAdapter) DoInTransaction(f func() error) error {
	f()

	tx, err := g.pgDb.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	//for _, action := range g.insertActions {
	//	for i := range action.Values {
	//		argsPlaceHolders := make([]string, len(action.Columns))
	//		for i := range action.Columns {
	//			argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	//		}
	//
	//		batch.Queue(
	//			fmt.Sprintf("insert into %s(%s) Values(%s)", action.TableName, strings.Join(action.Columns, ","), strings.Join(argsPlaceHolders, ",")),
	//			action.Values[i]...
	//		)
	//	}
	//}

	br := tx.SendBatch(context.Background(), batch)

	defer br.Close()

	if _, err = br.Exec(); err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
