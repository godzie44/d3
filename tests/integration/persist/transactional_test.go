package persist

import (
	"context"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
	"os"
	"sync"
	"testing"
)

type TransactionalTs struct {
	suite.Suite
	pgConn    *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
}

func (t *TransactionalTs) SetupSuite() {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	t.NoError(err)

	t.pgConn = driver.UnwrapConn().(*pgx.Conn)

	err = createSchema(t.pgConn)
	t.NoError(err)

	t.dbAdapter = helpers.NewDbAdapterWithQueryCounter(driver)
	t.d3Orm = orm.New(t.dbAdapter)
	t.NoError(t.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	t.NoError(err)
}

func (t *TransactionalTs) TearDownSuite() {
	t.NoError(deleteSchema(t.pgConn))
}

func (t *TransactionalTs) TearDownTest() {
	t.dbAdapter.ResetCounters()
	t.NoError(clearSchema(t.pgConn))
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

	helpers.NewPgTester(t.T(), t.pgConn).
		SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func newConn() *pgx.Conn {
	newConn, _ := pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	return newConn
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

	pgTester := helpers.NewPgTester(t.T(), newConn())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.CommitTx())

	pgTester.SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
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

	pgTester := helpers.NewPgTester(t.T(), newConn())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.RollbackTx())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func TestTransactionalTs(t *testing.T) {
	suite.Run(t, new(TransactionalTs))
}

type MultipleTransactionTs struct {
	suite.Suite
	pgConn    *pgxpool.Pool
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
}

func (t *MultipleTransactionTs) SetupSuite() {
	cfg, _ := pgxpool.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxPoolDriver(cfg)
	t.NoError(err)

	t.d3Orm = orm.New(driver)
	t.NoError(t.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	t.pgConn = driver.UnwrapConn().(*pgxpool.Pool)

	_, err = t.pgConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop_p(
		id SERIAL PRIMARY KEY,
		profile_id integer,
		name character varying(200) NOT NULL
	)`)
	t.NoError(err)
	_, err = t.pgConn.Exec(context.Background(), `INSERT INTO shop_p(id, name) VALUES (1, 'shop')`)
	t.NoError(err)
}

func (t *MultipleTransactionTs) TearDownSuite() {
	_, err := t.pgConn.Exec(context.Background(), `DROP TABLE shop_p`)
	t.NoError(err)
}

func (t *MultipleTransactionTs) TestConcurrentTransactionQuerying() {
	repository, _ := t.d3Orm.MakeRepository((*Shop)(nil))

	syncChan := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		ctx := t.d3Orm.CtxWithSession(context.Background())
		session := orm.Session(ctx)
		t.NoError(session.BeginTx())

		shop, _ := repository.FindOne(ctx, repository.Select().Where("id", "=", 1))
		shop.(*Shop).Name = "changed name"
		t.NoError(session.Flush())

		syncChan <- struct{}{}
		<-syncChan

		sameShop, _ := repository.FindOne(ctx, repository.Select().Where("name", "=", "changed name"))
		t.NotNil(sameShop)

		t.NoError(session.CommitTx())
	}()

	go func() {
		defer wg.Done()
		ctx := t.d3Orm.CtxWithSession(context.Background())
		session := orm.Session(ctx)
		t.NoError(session.BeginTx())

		<-syncChan

		shop, _ := repository.FindOne(ctx, repository.Select().Where("id", "=", 1))

		t.Equal("shop", shop.(*Shop).Name)
		syncChan <- struct{}{}
		t.NoError(session.CommitTx())
	}()

	wg.Wait()
}

func TestMultipleTransactionalTs(t *testing.T) {
	suite.Run(t, new(MultipleTransactionTs))
}
