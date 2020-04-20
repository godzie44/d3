package helpers

import (
	"context"
	"d3/orm"
	"d3/orm/persistence"
	"d3/orm/query"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type dbAdapterWithQueryCounter struct {
	queryCounter, insertCounter, updateCounter, deleteCounter int
	dbAdapter                                                 orm.Storage
}

func NewDbAdapterWithQueryCounter(dbAdapter orm.Storage) *dbAdapterWithQueryCounter {
	wrappedAdapter := &dbAdapterWithQueryCounter{dbAdapter: dbAdapter}

	dbAdapter.AfterQuery(func(_ string, _ ...interface{}) {
		wrappedAdapter.queryCounter++
	})

	return wrappedAdapter
}

func (d *dbAdapterWithQueryCounter) Insert(table string, cols, pkCols []string, values []interface{}, propagatePk bool, propagationFn func(scanner persistence.Scanner) error) error {
	d.insertCounter++
	return d.dbAdapter.Insert(table, cols, pkCols, values, propagatePk, propagationFn)
}

func (d *dbAdapterWithQueryCounter) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
	d.updateCounter++
	return d.dbAdapter.Update(table, cols, values, identityCond)
}

func (d *dbAdapterWithQueryCounter) Remove(table string, identityCond map[string]interface{}) error {
	d.deleteCounter++
	return d.dbAdapter.Remove(table, identityCond)
}

func (d *dbAdapterWithQueryCounter) ExecuteQuery(query *query.Query) ([]map[string]interface{}, error) {
	return d.dbAdapter.ExecuteQuery(query)
}

func (d *dbAdapterWithQueryCounter) BeforeQuery(fn func(query string, args ...interface{})) {
	d.dbAdapter.BeforeQuery(fn)
}

func (d *dbAdapterWithQueryCounter) AfterQuery(fn func(query string, args ...interface{})) {
	d.dbAdapter.AfterQuery(fn)
}

func (d *dbAdapterWithQueryCounter) QueryCounter() int {
	return d.queryCounter
}

func (d *dbAdapterWithQueryCounter) UpdateCounter() int {
	return d.updateCounter
}

func (d *dbAdapterWithQueryCounter) InsertCounter() int {
	return d.insertCounter
}

func (d *dbAdapterWithQueryCounter) DeleteCounter() int {
	return d.deleteCounter
}

type pgTester struct {
	t    *testing.T
	conn *pgx.Conn
}

func NewPgTester(t *testing.T, conn *pgx.Conn) *pgTester {
	return &pgTester{t, conn}
}

func (p *pgTester) SeeOne(sql string, args ...interface{}) *pgTester {
	return p.See(1, sql, args...)
}

func (p *pgTester) SeeTwo(sql string, args ...interface{}) *pgTester {
	return p.See(2, sql, args...)
}

func (p *pgTester) SeeThree(sql string, args ...interface{}) *pgTester {
	return p.See(3, sql, args...)
}

func (p *pgTester) SeeFour(sql string, args ...interface{}) *pgTester {
	return p.See(4, sql, args...)
}

func (p *pgTester) See(count int, sql string, args ...interface{}) *pgTester {
	var cnt int
	err := p.conn.QueryRow(context.Background(), fmt.Sprintf("SELECT count(*) cnt FROM (%s) t", sql), args...).Scan(&cnt)
	assert.NoError(p.t, err)

	assert.Equal(p.t, count, cnt)

	return p
}
