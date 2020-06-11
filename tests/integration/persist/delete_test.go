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

type DeleteTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	ctx       context.Context
}

func (d *DeleteTS) SetupSuite() {
	d.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	err := createSchema(d.pgDb)

	d.dbAdapter = helpers.NewDbAdapterWithQueryCounter(pgx2.NewGoPgXAdapter(d.pgDb, &adapter.SquirrelAdapter{}))
	d.d3Orm = orm.NewOrm(d.dbAdapter)
	d.NoError(d.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	d.NoError(err)
}

func (d *DeleteTS) SetupTest() {
	d.ctx = d.d3Orm.CtxWithSession(context.Background())
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

	rep, err := d.d3Orm.MakeRepository((*ShopProfile)(nil))
	d.NoError(err)

	profile, err := rep.FindOne(d.ctx, rep.MakeQuery().AndWhere("profile_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(d.ctx, profile))

	d.NoError(orm.Session(d.ctx).Flush())

	d.Equal(1, d.dbAdapter.DeleteCounter())
}

func (d *DeleteTS) TestDeleteWithRelations() {
	fillDb(d.Assert(), d.dbAdapter)

	rep, err := d.d3Orm.MakeRepository((*Shop)(nil))
	d.NoError(err)

	shop, err := rep.FindOne(d.ctx, rep.MakeQuery().AndWhere("shop_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(d.ctx, shop))

	d.dbAdapter.ResetCounters()
	d.NoError(orm.Session(d.ctx).Flush())

	// delete shop and profile (cause cascade)
	d.Equal(2, d.dbAdapter.DeleteCounter())

	// set books shop_id attribute to null where book_id = shop.ID (cause nullable)
	d.Equal(1, d.dbAdapter.UpdateCounter())
}

func (d *DeleteTS) TestDeleteWithManyToMany() {
	fillDb(d.Assert(), d.dbAdapter)

	rep, err := d.d3Orm.MakeRepository((*Book)(nil))
	d.NoError(err)

	book, err := rep.FindOne(d.ctx, rep.MakeQuery().AndWhere("book_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(d.ctx, book))

	d.dbAdapter.ResetCounters()
	d.NoError(orm.Session(d.ctx).Flush())

	// delete from book_p table and book_author_p table
	d.Equal(2, d.dbAdapter.DeleteCounter())
}

func TestDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteTS))
}
