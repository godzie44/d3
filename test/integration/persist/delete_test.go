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
	d.session = d.d3Orm.MakeSession()
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

	rep, err := d.session.MakeRepository((*ShopProfile)(nil))
	d.NoError(err)

	profile, err := rep.FindOne(rep.CreateQuery().AndWhere("profile_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(profile))

	d.NoError(d.session.Flush())

	d.Equal(1, d.dbAdapter.DeleteCounter())
}

func (d *DeleteTS) TestDeleteWithRelations() {
	fillDb(d.Assert(), d.dbAdapter)

	rep, err := d.session.MakeRepository((*Shop)(nil))
	d.NoError(err)

	shop, err := rep.FindOne(rep.CreateQuery().AndWhere("shop_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(shop))

	d.dbAdapter.ResetCounters()
	d.NoError(d.session.Flush())

	// delete shop and profile (cause cascade)
	d.Equal(2, d.dbAdapter.DeleteCounter())

	// set books shop_id attribute to null where book_id = shop.ID (cause nullable)
	d.Equal(1, d.dbAdapter.UpdateCounter())
}

func (d *DeleteTS) TestDeleteWithManyToMany() {
	fillDb(d.Assert(), d.dbAdapter)

	rep, err := d.session.MakeRepository((*Book)(nil))
	d.NoError(err)

	book, err := rep.FindOne(rep.CreateQuery().AndWhere("book_p.id = 1001"))
	d.NoError(err)

	d.NoError(rep.Delete(book))

	d.dbAdapter.ResetCounters()
	d.NoError(d.session.Flush())

	// delete from book_p table and book_author_p table
	d.Equal(2, d.dbAdapter.DeleteCounter())
}

func TestDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteTS))
}
