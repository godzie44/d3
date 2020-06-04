package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/tests/helpers"
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
	session   *orm.Session
}

func (t *TransactionalTs) SetupSuite() {
	t.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	err := createSchema(t.pgDb)

	t.dbAdapter = helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(t.pgDb, &adapter.SquirrelAdapter{}))
	t.d3Orm = orm.NewOrm(t.dbAdapter)
	t.NoError(t.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	t.NoError(err)
}

func (t *TransactionalTs) SetupTest() {
	t.session = t.d3Orm.MakeSession()
}

func (t *TransactionalTs) TearDownSuite() {
	t.NoError(deleteSchema(t.pgDb))
}

func (t *TransactionalTs) TearDownTest() {
	t.dbAdapter.ResetCounters()
	t.NoError(clearSchema(t.pgDb))
}

func (t *TransactionalTs) TestAutoCommit() {
	session := t.d3Orm.MakeSession()

	repository, _ := session.MakeRepository((*Shop)(nil))

	shop1 := &Shop{
		Name: "shop1",
	}
	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(shop1, shop2))
	t.NoError(session.Flush())

	helpers.NewPgTester(t.T(), t.pgDb).
		SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func newConn() *pgx.Conn {
	newConn, _ := pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	return newConn
}

func (t *TransactionalTs) TestManualCommit() {
	session := t.d3Orm.MakeSession()

	repository, _ := session.MakeRepository((*Shop)(nil))

	t.NoError(session.BeginTx())

	shop1 := &Shop{
		Name: "shop1",
	}
	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(shop1, shop2))
	t.NoError(session.Flush())

	pgTester := helpers.NewPgTester(t.T(), newConn())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.CommitTx())

	pgTester.SeeTwo("SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func (t *TransactionalTs) TestManualRollback() {
	session := t.d3Orm.MakeSession()

	repository, _ := session.MakeRepository((*Shop)(nil))

	t.NoError(session.BeginTx())

	shop1 := &Shop{
		Name: "shop1",
	}
	t.NoError(repository.Persists(shop1))
	t.NoError(session.Flush())

	shop2 := &Shop{
		Name: "shop2",
	}

	t.NoError(repository.Persists(shop2))
	t.NoError(session.Flush())

	pgTester := helpers.NewPgTester(t.T(), newConn())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")

	t.NoError(session.RollbackTx())

	pgTester.See(0, "SELECT * FROM shop_p WHERE name = $1 or name = $2", "shop1", "shop2")
}

func TestTransactionalTs(t *testing.T) {
	suite.Run(t, new(TransactionalTs))
}
