package persist

import (
	"context"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeleteTS struct {
	suite.Suite
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	ctx       context.Context
	execSqlFn func(sql string) error
}

func (d *DeleteTS) SetupSuite() {
	d.NoError(d.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	schemaSql, err := d.d3Orm.GenerateSchema()
	d.NoError(err)

	d.NoError(d.execSqlFn(schemaSql))
}

func (d *DeleteTS) SetupTest() {
	d.ctx = d.d3Orm.CtxWithSession(context.Background())
	fillDb(d.Assert(), d.dbAdapter)
}

func (d *DeleteTS) TearDownSuite() {
	d.NoError(d.execSqlFn(`
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`))
}

func (d *DeleteTS) TearDownTest() {
	d.dbAdapter.ResetCounters()
	d.NoError(d.execSqlFn(`
delete from book_p;
delete from author_p;
delete from book_author_p;
delete from shop_p;
delete from profile_p;
`))
}

func TestPGDeleteTestSuite(t *testing.T) {
	adapter, d3orm, execSqlFn, _ := db.CreatePGTestComponents(t)

	ts := &DeleteTS{
		dbAdapter: adapter,
		d3Orm:     d3orm,
		execSqlFn: execSqlFn,
	}
	suite.Run(t, ts)
}

func TestSQLiteDeleteTestSuite(t *testing.T) {
	adapter, d3orm, execSqlFn, _ := db.CreateSQLiteTestComponents(t, "_del")

	ts := &DeleteTS{
		d3Orm:     d3orm,
		dbAdapter: adapter,
		execSqlFn: execSqlFn,
	}

	suite.Run(t, ts)
}

func (d *DeleteTS) TestDeleteEntity() {
	rep, err := d.d3Orm.MakeRepository((*ShopProfile)(nil))
	d.NoError(err)

	profile, err := rep.FindOne(d.ctx, rep.Select().Where("profile_p.id", "=", 1001))
	d.NoError(err)

	d.NoError(rep.Delete(d.ctx, profile))

	d.NoError(orm.Session(d.ctx).Flush())

	d.Equal(1, d.dbAdapter.DeleteCounter())
}

func (d *DeleteTS) TestDeleteWithRelations() {
	rep, err := d.d3Orm.MakeRepository((*Shop)(nil))
	d.NoError(err)

	shop, err := rep.FindOne(d.ctx, rep.Select().Where("shop_p.id", "=", 1001))
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
	rep, err := d.d3Orm.MakeRepository((*Book)(nil))
	d.NoError(err)

	book, err := rep.FindOne(d.ctx, rep.Select().Where("book_p.id", "=", 1001))
	d.NoError(err)

	d.NoError(rep.Delete(d.ctx, book))

	d.dbAdapter.ResetCounters()
	d.NoError(orm.Session(d.ctx).Flush())

	// delete from book_p table and book_author_p table
	d.Equal(2, d.dbAdapter.DeleteCounter())
}
