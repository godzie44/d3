package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/test/helpers"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeleteTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	session   *orm.Session
}

func (d *DeleteTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	d.pgDb, _ = pgx.Connect(context.Background(), dsn)

	err := createSchema(d.pgDb)

	d.dbAdapter = helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(d.pgDb, &adapter.SquirrelAdapter{}))
	d.d3Orm = orm.NewOrm(d.dbAdapter)
	d.NoError(d.d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	d.NoError(err)
}

func (d *DeleteTS) SetupTest() {
	d.session = d.d3Orm.CreateSession()
}

func (d *DeleteTS) TearDownSuite() {
	d.NoError(deleteSchema(d.pgDb))
}

func (d *DeleteTS) TearDownTest() {
	d.dbAdapter.ResetCounters()
	d.NoError(clearSchema(d.pgDb))
}

func (d *DeleteTS) TestDeleteEntity() {
	fillDb(d.Assert(), d.dbAdapter)

	rep, err := d.d3Orm.CreateRepository(d.session, (*Shop)(nil))
	d.NoError(err)

	shop, err := rep.FindOne(rep.CreateQuery().AndWhere("shop_p.id = 1001"))

	d.NoError(rep.Delete(shop))

	d.NoError(d.session.Flush())

	d.Equal(1, d.dbAdapter.DeleteCounter())
}

func TestDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteTS))
}
