package persist

import (
	"context"
	"database/sql"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"sync"
	"testing"
)

type TransactionalTs struct {
	suite.Suite
	tester, independentTester helpers.DBTester
	execSqlFn                 func(sql string) error
	dbAdapter                 *helpers.DbAdapterWithQueryCounter
	d3Orm                     *orm.Orm
}

func (t *TransactionalTs) SetupSuite() {
	t.NoError(t.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	schemaSql, err := t.d3Orm.GenerateSchema()
	t.NoError(err)

	t.NoError(t.execSqlFn(schemaSql))
}

func (t *TransactionalTs) TearDownSuite() {
	t.NoError(t.execSqlFn(`
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`))
}

func (t *TransactionalTs) TearDownTest() {
	t.dbAdapter.ResetCounters()
	t.NoError(t.execSqlFn(`
delete from book_p;
delete from author_p;
delete from book_author_p;
delete from shop_p;
delete from profile_p;
`))
}

func TestPGTransactionalTs(t *testing.T) {
	adapter, d3orm, execSqlFn, tester := db.CreatePGTestComponents(t)

	indConn, _ := pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	ts := &TransactionalTs{
		dbAdapter:         adapter,
		d3Orm:             d3orm,
		execSqlFn:         execSqlFn,
		tester:            tester,
		independentTester: helpers.NewPgTester(t, indConn),
	}

	suite.Run(t, ts)
}

func TestSQLiteTransactionalTs(t *testing.T) {
	adapter, d3orm, execSqlFn, tester := db.CreateSQLiteTestComponents(t)

	indConn, _ := sql.Open("sqlite3", "./../../data/sqlite/test.db")

	ts := &TransactionalTs{
		d3Orm:             d3orm,
		dbAdapter:         adapter,
		execSqlFn:         execSqlFn,
		tester:            tester,
		independentTester: helpers.NewSQLiteTester(t, indConn),
	}

	suite.Run(t, ts)
}

func (t *TransactionalTs) TestAutoCommit() {
	ctx := t.d3Orm.CtxWithSession(context.Background())
	session := orm.Session(ctx)
	repository, _ := t.d3Orm.MakeRepository((*Shop)(nil))

	shop1 := &Shop{
		Name: "shop1",
	}
	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(ctx, shop1, shop2))
	t.NoError(session.Flush())

	t.tester.
		SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func (t *TransactionalTs) TestManualCommit() {
	ctx := t.d3Orm.CtxWithSession(context.Background())
	session := orm.Session(ctx)

	repository, _ := t.d3Orm.MakeRepository((*Shop)(nil))

	t.NoError(session.BeginTx())

	shop1 := &Shop{
		Name: "shop1",
	}
	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(ctx, shop1, shop2))
	t.NoError(session.Flush())

	t.independentTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.CommitTx())

	t.independentTester.SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func (t *TransactionalTs) TestManualRollback() {
	ctx := t.d3Orm.CtxWithSession(context.Background())
	session := orm.Session(ctx)

	repository, _ := t.d3Orm.MakeRepository((*Shop)(nil))

	t.NoError(session.BeginTx())

	shop1 := &Shop{
		Name: "shop1",
	}
	t.NoError(repository.Persists(ctx, shop1))
	t.NoError(session.Flush())

	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(ctx, shop2))
	t.NoError(session.Flush())

	t.independentTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.RollbackTx())

	t.independentTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

type MultipleTransactionTs struct {
	suite.Suite
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	execSqlFn func(sql string) error
}

func (m *MultipleTransactionTs) SetupSuite() {
	m.NoError(m.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	schemaSql, err := m.d3Orm.GenerateSchema()
	m.NoError(err)

	m.NoError(m.execSqlFn(schemaSql))

	m.NoError(m.execSqlFn(`INSERT INTO shop_p(id, name) VALUES (1, 'shop')`))
}

func (m *MultipleTransactionTs) TearDownSuite() {
	m.NoError(m.execSqlFn(`
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`))
}

func (m *MultipleTransactionTs) TestConcurrentTransactionQuerying() {
	repository, _ := m.d3Orm.MakeRepository((*Shop)(nil))

	syncChan := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		ctx := m.d3Orm.CtxWithSession(context.Background())
		session := orm.Session(ctx)
		m.NoError(session.BeginTx())

		shop, _ := repository.FindOne(ctx, repository.Select().Where("id", "=", 1))
		shop.(*Shop).Name = "changed name"
		m.NoError(session.Flush())

		syncChan <- struct{}{}
		<-syncChan

		sameShop, _ := repository.FindOne(ctx, repository.Select().Where("name", "=", "changed name"))
		m.NotNil(sameShop)

		m.NoError(session.CommitTx())
	}()

	go func() {
		defer wg.Done()
		ctx := m.d3Orm.CtxWithSession(context.Background())
		session := orm.Session(ctx)
		m.NoError(session.BeginTx())

		<-syncChan

		shop, _ := repository.FindOne(ctx, repository.Select().Where("id", "=", 1))

		m.Equal("shop", shop.(*Shop).Name)
		syncChan <- struct{}{}
		m.NoError(session.CommitTx())
	}()

	wg.Wait()
}

func TestPGMultipleTransactionalTs(t *testing.T) {
	cfg, _ := pgxpool.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxPoolDriver(cfg)
	assert.NoError(t, err)

	conn := driver.UnwrapConn().(*pgxpool.Pool)
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(driver)

	ts := &MultipleTransactionTs{
		dbAdapter: dbAdapter,
		d3Orm:     orm.New(dbAdapter),
		execSqlFn: func(sql string) error {
			_, err := conn.Exec(context.Background(), sql)
			return err
		},
	}
	suite.Run(t, ts)
}

func TestSQLiteMultipleTransactionalTs(t *testing.T) {
	adapter, d3orm, execSqlFn, _ := db.CreateSQLiteTestComponents(t)

	ts := &MultipleTransactionTs{
		d3Orm:     d3orm,
		dbAdapter: adapter,
		execSqlFn: execSqlFn,
	}
	suite.Run(t, ts)
}
