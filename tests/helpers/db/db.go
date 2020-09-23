package db

import (
	"context"
	"database/sql"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/adapter/sqlite"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func CreatePGTestComponents(t *testing.T) (*helpers.DbAdapterWithQueryCounter, *orm.Orm, func(sql string) error, helpers.DBTester) {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	assert.NoError(t, err)

	conn := driver.UnwrapConn().(*pgx.Conn)

	dbAdapter := helpers.NewDbAdapterWithQueryCounter(driver)

	execSqlFn := func(sql string) error {
		_, err := conn.Exec(context.Background(), sql)
		return err
	}

	return dbAdapter, orm.New(dbAdapter), execSqlFn, helpers.NewPgTester(t, conn)
}

func CreateSQLiteTestComponents(t *testing.T, postfix string) (*helpers.DbAdapterWithQueryCounter, *orm.Orm, func(sql string) error, helpers.DBTester) {
	driver, err := sqlite.NewSQLiteDriver("./../../data/sqlite/test" + postfix + ".db")
	assert.NoError(t, err)

	conn := driver.UnwrapConn().(*sql.DB)
	d := helpers.NewDbAdapterWithQueryCounter(driver)

	execSqlFn := func(sql string) error {
		_, err := conn.Exec(sql)
		return err
	}

	return d, orm.New(d), execSqlFn, helpers.NewSQLiteTester(t, conn)
}
