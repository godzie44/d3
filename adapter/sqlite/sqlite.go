package sqlite

import (
	"errors"
	"fmt"
	"github.com/godzie44/d3/adapter"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/orm/schema"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"strconv"
	"strings"

	"database/sql"
)

type sqliteDriver struct {
	beforeQCallback, afterQCallback []func(query string, args ...interface{})

	db *sql.DB
}

func NewSQLiteDriver(dataSourceName string) (*sqliteDriver, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &sqliteDriver{
		db: db,
	}, nil
}

func (s *sqliteDriver) UnwrapConn() interface{} {
	return s.db
}

func (s *sqliteDriver) Close() error {
	return s.db.Close()
}

func (s *sqliteDriver) ExecuteQuery(query *query.Query, tx orm.Transaction) ([]map[string]interface{}, error) {
	q, args, err := adapter.QueryToSql(query)
	if err != nil {
		return nil, err
	}

	for i := range s.beforeQCallback {
		s.beforeQCallback[i](q, args...)
	}

	sqliteTx, ok := tx.(*sqliteTransaction)
	if !ok {
		panic(errors.New("transaction type must be sqliteTransaction"))
	}

	rows, err := sqliteTx.tx.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := range s.afterQCallback {
		s.afterQCallback[i](q, args...)
	}

	result := make([]map[string]interface{}, 0)

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = values[i]
		}

		result = append(result, m)
	}

	return result, nil
}

func (s *sqliteDriver) BeforeQuery(fn func(query string, args ...interface{})) {
	s.beforeQCallback = append(s.beforeQCallback, fn)
}

func (s *sqliteDriver) AfterQuery(fn func(query string, args ...interface{})) {
	s.afterQCallback = append(s.afterQCallback, fn)
}

func (s *sqliteDriver) MakeScalarDataMapper() orm.ScalarDataMapper {
	return func(data interface{}, into reflect.Kind) interface{} {
		switch into {
		case reflect.Int32:
			return int32(data.(int64))
		case reflect.Int:
			return int(data.(int64))
		case reflect.Float32:
			return float32(data.(float64))

		default:
			return data
		}
	}
}

type sqliteTransaction struct {
	tx *sql.Tx
}

func (t *sqliteTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTransaction) Rollback() error {
	return t.tx.Rollback()
}

func (s *sqliteDriver) BeginTx() (orm.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &sqliteTransaction{
		tx: tx,
	}, nil
}

func (s *sqliteDriver) MakePusher(tx orm.Transaction) persistence.Pusher {
	sqliteTx, ok := tx.(*sqliteTransaction)
	if !ok {
		panic(errors.New("transaction type must be sqliteTransaction"))
	}

	return &sqlitePusher{tx: sqliteTx}
}

type sqlitePusher struct {
	tx *sqliteTransaction
}

func (s *sqlitePusher) Insert(table string, cols []string, values []interface{}, onConflict persistence.OnConflict) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	var onConflictClause string
	switch onConflict {
	case persistence.DoNothing:
		onConflictClause = "OR IGNORE"
	case persistence.Undefined:
		break
	}

	sql := fmt.Sprintf("INSERT %s INTO %s(%s) VALUES(%s)", onConflictClause, table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ","))

	_, err := s.tx.tx.Exec(
		sql,
		values...,
	)
	if err != nil {
		return fmt.Errorf("insert sqlite driver: %w", err)
	}

	return nil
}

type idScanner struct {
	id int64
}

func (i *idScanner) Scan(v ...interface{}) error {
	if len(v) > 0 {
		switch v[0].(type) {
		case *sql.NullInt32:
			return v[0].(*sql.NullInt32).Scan(int32(i.id))
		case *sql.NullInt64:
			return v[0].(*sql.NullInt64).Scan(i.id)
		default:
			return errors.New("unknown pk type, expected sql.NullInt32, sql.NullInt64")
		}
	}
	return nil
}

func (s *sqlitePusher) InsertWithReturn(
	table string,
	cols []string,
	values []interface{},
	_ []string,
	withReturned func(scanner persistence.Scanner) error,
) error {
	argsPlaceHolders := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		argsPlaceHolders[i] = "$" + strconv.Itoa(i+1)
	}

	res, err := s.tx.tx.Exec(
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(cols, ","), strings.Join(argsPlaceHolders, ",")),
		values...,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if err := withReturned(&idScanner{id: id}); err != nil {
		return fmt.Errorf("insert sqlite driver: %w", err)
	}

	return nil
}

func (s *sqlitePusher) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
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

	_, err := s.tx.tx.Exec(
		fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setCommands, ","), whereStr),
		queryValues...,
	)
	if err != nil {
		return fmt.Errorf("insert sqlite driver: %w", err)
	}

	return nil
}

func (s *sqlitePusher) Remove(table string, identityCond map[string]interface{}) error {
	args := make([]interface{}, 0, len(identityCond))
	where := make([]string, 0, len(identityCond))

	for col, val := range identityCond {
		args = append(args, val)
		where = append(where, col+"=$"+strconv.Itoa(len(args)))
	}

	_, err := s.tx.tx.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(where, " AND ")),
		args...,
	)

	return err
}

func (s *sqliteDriver) CreateTableSql(name string, columns map[string]schema.ColumnType, pkColumns []string, pkStrategy entity.PkStrategy) string {
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
		needAI := false

		colSql := strings.Builder{}
		colSql.WriteString(col)
		colSql.WriteRune(' ')

		switch ctype {
		case schema.UUID:
			colSql.WriteString("TEXT")
		case schema.Bool:
			colSql.WriteString("BOOLEAN NOT NULL")
		case schema.Int:
			if isPkCol(col) && pkStrategy == entity.Auto {
				needAI = true
				colSql.WriteString("INTEGER NOT NULL")
			} else {
				colSql.WriteString("BIGINT NOT NULL")
			}
		case schema.Int32:
			if isPkCol(col) && pkStrategy == entity.Auto {
				needAI = true
			}
			colSql.WriteString("INTEGER NOT NULL")
		case schema.Int64:
			if isPkCol(col) && pkStrategy == entity.Auto {
				needAI = true
				colSql.WriteString("INTEGER NOT NULL")
			} else {
				colSql.WriteString("BIGINT NOT NULL")
			}
		case schema.Float32:
			colSql.WriteString("FLOAT NOT NULL")
		case schema.Float64:
			colSql.WriteString("DOUBLE NOT NULL")
		case schema.String:
			colSql.WriteString("TEXT NOT NULL")
		case schema.Time:
			colSql.WriteString("datetime NOT NULL")
		case schema.NullBool:
			colSql.WriteString("BOOLEAN")
		case schema.NullInt64:
			if isPkCol(col) && pkStrategy == entity.Auto {
				needAI = true
				colSql.WriteString("INTEGER")
			} else {
				colSql.WriteString("BIGINT")
			}
		case schema.NullInt32:
			if isPkCol(col) && pkStrategy == entity.Auto {
				needAI = true
			}
			colSql.WriteString("INTEGER")
		case schema.NullFloat64:
			colSql.WriteString("DOUBLE")
		case schema.NullString:
			colSql.WriteString("TEXT")
		case schema.NullTime:
			colSql.WriteString("datetime")
		}

		if isPkCol(col) && len(pkColumns) == 1 {
			colSql.WriteRune(' ')
			colSql.WriteString("PRIMARY KEY")
			if needAI {
				colSql.WriteString(" AUTOINCREMENT")
			}
		}

		colsSql = append(colsSql, colSql.String())
	}
	sql.WriteString(strings.Join(colsSql, ",\n"))

	if len(pkColumns) > 1 {
		sql.WriteString(",\n")
		sql.WriteString(fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkColumns, ",")))
	}

	sql.WriteString("\n);\n")

	return sql.String()
}

func (s *sqliteDriver) CreateIndexSql(name string, table string, columns ...string) string {
	panic("implement me")
}
