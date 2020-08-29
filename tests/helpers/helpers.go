package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/orm/schema"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type DbAdapterWithQueryCounter struct {
	queryCounter, insertCounter, updateCounter, deleteCounter int
	dbAdapter                                                 orm.Driver
}

func (d *DbAdapterWithQueryCounter) CreateTableSql(name string, columns map[string]schema.ColumnType, pkColumns []string, pkStrategy entity.PkStrategy) string {
	if generator, ok := d.dbAdapter.(schema.StorageSchemaGenerator); ok {
		return generator.CreateTableSql(name, columns, pkColumns, pkStrategy)
	}
	panic("not implemented")
}

func (d *DbAdapterWithQueryCounter) CreateIndexSql(name string, table1 string, columns ...string) string {
	if generator, ok := d.dbAdapter.(schema.StorageSchemaGenerator); ok {
		return generator.CreateIndexSql(name, table1, columns...)
	}
	panic("not implemented")
}

func (d *DbAdapterWithQueryCounter) MakeScalarDataMapper() orm.ScalarDataMapper {
	return d.dbAdapter.MakeScalarDataMapper()
}

func (d *DbAdapterWithQueryCounter) MakePusher(tx orm.Transaction) persistence.Pusher {
	ps := d.dbAdapter.MakePusher(tx)
	return &persistStoreWithCounters{
		ps: ps,
		onInsert: func() {
			d.insertCounter++
		},
		onUpdate: func() {
			d.updateCounter++
		},
		onDelete: func() {
			d.deleteCounter++
		},
	}
}

func (d *DbAdapterWithQueryCounter) BeginTx() (orm.Transaction, error) {
	return d.dbAdapter.BeginTx()
}

func NewDbAdapterWithQueryCounter(dbAdapter orm.Driver) *DbAdapterWithQueryCounter {
	wrappedAdapter := &DbAdapterWithQueryCounter{dbAdapter: dbAdapter}

	dbAdapter.AfterQuery(func(_ string, _ ...interface{}) {
		wrappedAdapter.queryCounter++
	})

	return wrappedAdapter
}

func (d *DbAdapterWithQueryCounter) ExecuteQuery(query *query.Query, tx orm.Transaction) ([]map[string]interface{}, error) {
	return d.dbAdapter.ExecuteQuery(query, tx)
}

func (d *DbAdapterWithQueryCounter) BeforeQuery(fn func(query string, args ...interface{})) {
	d.dbAdapter.BeforeQuery(fn)
}

func (d *DbAdapterWithQueryCounter) AfterQuery(fn func(query string, args ...interface{})) {
	d.dbAdapter.AfterQuery(fn)
}

func (d *DbAdapterWithQueryCounter) QueryCounter() int {
	return d.queryCounter
}

func (d *DbAdapterWithQueryCounter) UpdateCounter() int {
	return d.updateCounter
}

func (d *DbAdapterWithQueryCounter) InsertCounter() int {
	return d.insertCounter
}

func (d *DbAdapterWithQueryCounter) DeleteCounter() int {
	return d.deleteCounter
}
func (d *DbAdapterWithQueryCounter) ResetCounters() {
	d.deleteCounter = 0
	d.updateCounter = 0
	d.insertCounter = 0
	d.queryCounter = 0
}

type persistStoreWithCounters struct {
	ps                           persistence.Pusher
	onInsert, onUpdate, onDelete func()
}

func (p *persistStoreWithCounters) Insert(table string, cols []string, values []interface{}, onConflict persistence.OnConflict) error {
	p.onInsert()
	return p.ps.Insert(table, cols, values, onConflict)
}

func (p *persistStoreWithCounters) InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner persistence.Scanner) error) error {
	p.onInsert()
	return p.ps.InsertWithReturn(table, cols, values, returnCols, withReturned)
}

func (p *persistStoreWithCounters) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
	p.onUpdate()
	return p.ps.Update(table, cols, values, identityCond)
}

func (p *persistStoreWithCounters) Remove(table string, identityCond map[string]interface{}) error {
	p.onDelete()
	return p.ps.Remove(table, identityCond)
}

type DBTester interface {
	SeeOne(sql string, args ...interface{}) DBTester
	SeeTwo(sql string, args ...interface{}) DBTester
	SeeThree(sql string, args ...interface{}) DBTester
	SeeFour(sql string, args ...interface{}) DBTester
	See(count int, sql string, args ...interface{}) DBTester
	SeeTable(tableName string) DBTester
}

type PgTester struct {
	t    *testing.T
	Conn *pgx.Conn
}

func NewPgTester(t *testing.T, conn *pgx.Conn) *PgTester {
	return &PgTester{t, conn}
}

func (p *PgTester) SeeOne(sql string, args ...interface{}) DBTester {
	return p.See(1, sql, args...)
}

func (p *PgTester) SeeTwo(sql string, args ...interface{}) DBTester {
	return p.See(2, sql, args...)
}

func (p *PgTester) SeeThree(sql string, args ...interface{}) DBTester {
	return p.See(3, sql, args...)
}

func (p *PgTester) SeeFour(sql string, args ...interface{}) DBTester {
	return p.See(4, sql, args...)
}

func (p *PgTester) See(count int, sql string, args ...interface{}) DBTester {
	var cnt int
	err := p.Conn.QueryRow(context.Background(), fmt.Sprintf("SELECT count(*) cnt FROM (%s) t", sql), args...).Scan(&cnt)
	assert.NoError(p.t, err)

	assert.Equal(p.t, count, cnt)
	return p
}

func (p *PgTester) SeeTable(tableName string) DBTester {
	var tableSql = "SELECT * FROM pg_tables where schemaname = 'public' and tablename=$1"

	var cnt int
	err := p.Conn.QueryRow(context.Background(), fmt.Sprintf("SELECT count(*) cnt FROM (%s) t", tableSql), tableName).Scan(&cnt)
	assert.NoError(p.t, err)

	assert.GreaterOrEqual(p.t, cnt, 1)
	return p
}

type SqliteTester struct {
	t    *testing.T
	Conn *sql.DB
}

func NewSQLiteTester(t *testing.T, conn *sql.DB) *SqliteTester {
	return &SqliteTester{t, conn}
}

func (s *SqliteTester) SeeOne(sql string, args ...interface{}) DBTester {
	return s.See(1, sql, args...)
}

func (s *SqliteTester) SeeTwo(sql string, args ...interface{}) DBTester {
	return s.See(2, sql, args...)
}

func (s *SqliteTester) SeeThree(sql string, args ...interface{}) DBTester {
	return s.See(3, sql, args...)
}

func (s *SqliteTester) SeeFour(sql string, args ...interface{}) DBTester {
	return s.See(4, sql, args...)
}

func (s *SqliteTester) See(count int, sql string, args ...interface{}) DBTester {
	var cnt int
	err := s.Conn.QueryRow(fmt.Sprintf("SELECT count(*) cnt FROM (%s) t", sql), args...).Scan(&cnt)
	assert.NoError(s.t, err)

	assert.Equal(s.t, count, cnt)
	return s
}

func (s *SqliteTester) SeeTable(tableName string) DBTester {
	var tableSql = "SELECT name FROM sqlite_master WHERE type='table' AND name=$1"

	var cnt int
	err := s.Conn.QueryRow(fmt.Sprintf("SELECT count(*) cnt FROM (%s) t", tableSql), tableName).Scan(&cnt)
	assert.NoError(s.t, err)

	assert.GreaterOrEqual(s.t, cnt, 1)
	return s
}
