package pgx

import (
	"context"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type PgxDriverTS struct {
	suite.Suite
	driver *pgxDriver
	tester *helpers.PgTester
}

type pgxExecer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func (p *PgxDriverTS) SetupTest() {
	_, err := p.driver.UnwrapConn().(pgxExecer).Exec(context.Background(), `CREATE TABLE IF NOT EXISTS pgx_test_table(
		id integer NOT NULL,
		data text NOT NULL,
		CONSTRAINT pgx_test_table_pk PRIMARY KEY (id)
	)`)
	p.NoError(err)

	_, err = p.driver.UnwrapConn().(pgxExecer).Exec(context.Background(), `
INSERT INTO pgx_test_table(id, data) VALUES (1, 'test 1');
INSERT INTO pgx_test_table(id, data) VALUES (2, 'test 2');
INSERT INTO pgx_test_table(id, data) VALUES (3, 'test 3');
`)
	p.NoError(err)
}

func (p *PgxDriverTS) TearDownTest() {
	_, err := p.driver.UnwrapConn().(pgxExecer).Exec(context.Background(), `DROP TABLE pgx_test_table;`)
	p.NoError(err)
}

func TestPgxConnDriverTestSuite(t *testing.T) {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := NewPgxDriver(cfg)
	assert.NoError(t, err)

	suite.Run(t, &PgxDriverTS{
		driver: driver,
		tester: helpers.NewPgTester(t, driver.UnwrapConn().(*pgx.Conn)),
	})
}

func TestPgxPoolDriverTestSuite(t *testing.T) {
	cfg, _ := pgxpool.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := NewPgxPoolDriver(cfg)
	assert.NoError(t, err)

	testerConn, err := driver.UnwrapConn().(*pgxpool.Pool).Acquire(context.Background())
	assert.NoError(t, err)

	suite.Run(t, &PgxDriverTS{
		driver: driver,
		tester: helpers.NewPgTester(t, testerConn.Conn()),
	})
}

func (p *PgxDriverTS) TestPgxDriverQuery() {
	data, err := p.driver.ExecuteQuery(query.New().Select("*").From("pgx_test_table").Where("id", "=", "1"))
	p.NoError(err)

	p.Len(data, 1)
	p.Equal(data[0], map[string]interface{}{"id": int32(1), "data": "test 1"})

	data, err = p.driver.ExecuteQuery(query.New().Select("*").From("pgx_test_table"))
	p.NoError(err)

	p.Len(data, 3)
}

func (p *PgxDriverTS) TestPgxDriverTxInsert() {
	tx, err := p.driver.BeginTx()
	p.NoError(err)

	pusher := p.driver.MakePusher(tx)

	err = pusher.Insert("pgx_test_table", []string{"id", "data"}, []interface{}{4, "test 4"}, persistence.Undefined)
	p.NoError(err)
	p.NoError(tx.Commit())

	p.tester.SeeFour("select * from pgx_test_table")
}

func (p *PgxDriverTS) TestPgxDriverTxUpdate() {
	tx, err := p.driver.BeginTx()
	p.NoError(err)

	pusher := p.driver.MakePusher(tx)

	err = pusher.Update("pgx_test_table", []string{"data"}, []interface{}{"test upd"}, map[string]interface{}{"id": 1})
	p.NoError(err)
	p.NoError(tx.Commit())

	p.tester.SeeOne("select * from pgx_test_table where data = 'test upd'")
}

func (p *PgxDriverTS) TestPgxDriverTxDelete() {
	tx, err := p.driver.BeginTx()
	p.NoError(err)

	pusher := p.driver.MakePusher(tx)

	err = pusher.Remove("pgx_test_table", map[string]interface{}{"id": 1})
	p.NoError(err)
	p.NoError(tx.Commit())

	p.tester.SeeTwo("select * from pgx_test_table")
}

func (p *PgxDriverTS) TestPgxDriverTxRollback() {
	tx, err := p.driver.BeginTx()
	p.NoError(err)

	pusher := p.driver.MakePusher(tx)

	err = pusher.Insert("pgx_test_table", []string{"id", "data"}, []interface{}{4, "test 4"}, persistence.Undefined)
	p.NoError(err)
	p.NoError(tx.Rollback())

	p.tester.SeeThree("select * from pgx_test_table")
}
