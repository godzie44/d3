package adapter

import (
	"context"
	"errors"
	"fmt"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/orm/schema"
	"github.com/jackc/pgx/v4"
	"reflect"
	"strconv"
	"strings"
)

type GoPgXAdapter struct {
	pgDb         *pgx.Conn
	queryAdapter *SquirrelAdapter

	beforeQCallback, afterQCallback []func(query string, args ...interface{})
}

func (g *GoPgXAdapter) MakeRawDataMapper() orm.RawDataMapper {
	return func(data interface{}, into reflect.Kind) interface{} {
		switch into {
		case reflect.Int:
			return int(data.(int64))
		default:
			return data
		}
	}
}

func (g *GoPgXAdapter) CreateTableSql(name string, columns map[string]schema.ColumnType, pkColumns []string, pkStrategy entity.PkStrategy) string {
	isPkCol := func(colName string) bool {
		for _, pkCol := range pkColumns {
			if pkCol == colName {
				return true
			}
		}
		return false
	}

	sql := strings.Builder{}
	sql.WriteString("CREATE TABLE IF NOT EXISTS " + name)
	sql.WriteString("(\n")

	var colsSql []string
	for col, ctype := range columns {
		colSql := strings.Builder{}
		colSql.WriteString(col)
		colSql.WriteRune(' ')

		switch ctype {
		case schema.Bool:
			colSql.WriteString("BOOLEAN NOT NULL")
		case schema.Int:
			if isPkCol(col) && pkStrategy == entity.Auto {
				colSql.WriteString("BIGSERIAL")
			} else {
				colSql.WriteString("BIGINT NOT NULL")
			}
		case schema.Int32:
			if isPkCol(col) && pkStrategy == entity.Auto {
				colSql.WriteString("SERIAL")
			} else {
				colSql.WriteString("INTEGER NOT NULL")
			}
		case schema.Int64:
			if isPkCol(col) && pkStrategy == entity.Auto {
				colSql.WriteString("BIGSERIAL")
			} else {
				colSql.WriteString("BIGINT NOT NULL")
			}
		case schema.Float32:
			colSql.WriteString("REAL NOT NULL")
		case schema.Float64:
			colSql.WriteString("DOUBLE PRECISION NOT NULL")
		case schema.String:
			colSql.WriteString("TEXT NOT NULL")
		case schema.Time:
			colSql.WriteString("TIMESTAMP WITH TIME ZONE NOT NULL")
		case schema.NullBool:
			colSql.WriteString("BOOLEAN")
		case schema.NullInt64:
			if isPkCol(col) && pkStrategy == entity.Auto {
				colSql.WriteString("BIGSERIAL")
			} else {
				colSql.WriteString("BIGINT")
			}
		case schema.NullInt32:
			if isPkCol(col) && pkStrategy == entity.Auto {
				colSql.WriteString("SERIAL")
			} else {
				colSql.WriteString("INTEGER")
			}
		case schema.NullFloat64:
			colSql.WriteString("DOUBLE PRECISION")
		case schema.NullString:
			colSql.WriteString("TEXT")
		case schema.NullTime:
			colSql.WriteString("TIMESTAMP WITH TIME ZONE")
		}

		if isPkCol(col) {
			colSql.WriteRune(' ')
			colSql.WriteString("PRIMARY KEY")
		}

		colsSql = append(colsSql, colSql.String())
	}
	sql.WriteString(strings.Join(colsSql, ",\n"))
	sql.WriteString("\n);\n")

	return sql.String()
}

func (g *GoPgXAdapter) CreateIndexSql(name string, table string, columns ...string) string {
	panic("implement me")
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

func (g *GoPgXAdapter) MakePusher(tx orm.Transaction) persistence.Pusher {
	pgxTx, ok := tx.(*pgxTransaction)
	if !ok {
		panic(errors.New("transaction type must be pgxTransaction"))
	}

	return &pgxPusher{tx: pgxTx}
}

type pgxPusher struct {
	tx *pgxTransaction
}

func (p *pgxPusher) Insert(table string, cols []string, values []interface{}) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	_, err := p.tx.tx.Exec(
		context.Background(),
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ",")),
		values...,
	)

	if err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}

	return nil
}

func (p *pgxPusher) InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner persistence.Scanner) error) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	row := p.tx.tx.QueryRow(
		context.Background(),
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ","), strings.Join(returnCols, ",")),
		values...,
	)

	if err := withReturned(row); err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}

	return nil
}

func (p *pgxPusher) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
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

	_, err := p.tx.tx.Exec(
		context.Background(),
		fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setCommands, ","), whereStr),
		queryValues...,
	)
	if err != nil {
		return fmt.Errorf("insert pgx driver: %w", err)
	}
	return nil
}

func (p *pgxPusher) Remove(table string, identityCond map[string]interface{}) error {
	args := make([]interface{}, 0, len(identityCond))
	where := make([]string, 0, len(identityCond))

	for col, val := range identityCond {
		args = append(args, val)
		where = append(where, col+"=$"+strconv.Itoa(len(args)))
	}

	_, err := p.tx.tx.Exec(
		context.Background(),
		fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(where, " AND ")),
		args...,
	)

	return err
}

type pgxTransaction struct {
	tx pgx.Tx
}

func (p *pgxTransaction) Commit() error {
	return p.tx.Commit(context.Background())
}

func (p *pgxTransaction) Rollback() error {
	return p.tx.Rollback(context.Background())
}

func (g *GoPgXAdapter) BeginTx() (orm.Transaction, error) {
	tx, err := g.pgDb.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	return &pgxTransaction{tx: tx}, nil
}
