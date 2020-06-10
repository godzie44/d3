package persist

import (
	"context"
	"github.com/godzie44/d3/adapter"
	pgx2 "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type TransactionalTs struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
}

func (t *TransactionalTs) SetupSuite() {
	t.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	err := createSchema(t.pgDb)

	t.dbAdapter = helpers.NewDbAdapterWithQueryCounter(pgx2.NewGoPgXAdapter(t.pgDb, &adapter.SquirrelAdapter{}))
	t.d3Orm = orm.NewOrm(t.dbAdapter)
	t.NoError(t.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	t.NoError(err)
}

func (t *TransactionalTs) TearDownSuite() {
	t.NoError(deleteSchema(t.pgDb))
}

func (t *TransactionalTs) TearDownTest() {
	t.dbAdapter.ResetCounters()
	t.NoError(clearSchema(t.pgDb))
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

	helpers.NewPgTester(t.T(), t.pgDb).
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
