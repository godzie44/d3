package persist

import (
	"context"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PersistsTS struct {
	suite.Suite
	tester    helpers.DBTester
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	ctx       context.Context
	execSqlFn func(sql string) error
}

func (o *PersistsTS) SetupSuite() {
	o.NoError(o.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	schemaSql, err := o.d3Orm.GenerateSchema()
	o.NoError(err)

	o.NoError(o.execSqlFn(schemaSql))
}

func (o *PersistsTS) SetupTest() {
	o.ctx = o.d3Orm.CtxWithSession(context.Background())
}

func (o *PersistsTS) TearDownSuite() {
	o.NoError(o.execSqlFn(`
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`))
}

func (o *PersistsTS) TearDownTest() {
	o.dbAdapter.ResetCounters()
	o.NoError(o.execSqlFn(`
delete from book_p;
delete from author_p;
delete from book_author_p;
delete from shop_p;
delete from profile_p;
`))
}

func TestPGPersistsSuite(t *testing.T) {
	adapter, d3orm, execSqlFn, tester := db.CreatePGTestComponents(t)

	ts := &PersistsTS{
		dbAdapter: adapter,
		d3Orm:     d3orm,
		execSqlFn: execSqlFn,
		tester:    tester,
	}
	suite.Run(t, ts)
}

func TestSQLitePersistsSuite(t *testing.T) {
	adapter, d3orm, execSqlFn, tester := db.CreateSQLiteTestComponents(t)

	ts := &PersistsTS{
		d3Orm:     d3orm,
		dbAdapter: adapter,
		execSqlFn: execSqlFn,
		tester:    tester,
	}
	suite.Run(t, ts)
}

func (o *PersistsTS) TestSimpleInsert() {
	repository, err := o.d3Orm.MakeRepository((*Shop)(nil))
	o.NoError(err)

	shop := &Shop{
		Books: nil,
		Profile: entity.NewCell(&ShopProfile{
			Description: "this is simple tests shop",
		}),
		Name: "simple-shop",
	}

	o.NoError(repository.Persists(o.ctx, shop))
	o.NoError(orm.Session(o.ctx).Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)

	o.tester.
		SeeOne("SELECT * FROM shop_p WHERE name='simple-shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is simple tests shop'")
}

func (o *PersistsTS) TestBigInsert() {
	shop, err := createAndPersistsShop(o.ctx, o.d3Orm)
	o.NoError(err)

	o.NoError(orm.Session(o.ctx).Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Authors.Get(0).(*Author).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Authors.Get(0).(*Author).Id.Int32)

	o.tester.
		SeeOne("SELECT * FROM shop_p WHERE name='shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is tests shop'").
		SeeTwo("SELECT * FROM book_p").
		SeeThree("SELECT * FROM author_p").
		SeeFour("SELECT * FROM book_author_p")
}

func (o *PersistsTS) TestNoNewQueriesIfDoubleFlush() {
	_, err := createAndPersistsShop(o.ctx, o.d3Orm)
	o.NoError(err)

	o.NoError(orm.Session(o.ctx).Flush())
	insertCounter, updCounter := o.dbAdapter.InsertCounter(), o.dbAdapter.UpdateCounter()

	o.NoError(orm.Session(o.ctx).Flush())

	o.Equal(insertCounter, o.dbAdapter.InsertCounter())
	o.Equal(updCounter, o.dbAdapter.UpdateCounter())
}

func (o *PersistsTS) TestOneNewEntityIfDoublePersist() {
	repository, _ := o.d3Orm.MakeRepository((*Shop)(nil))

	shop := &Shop{
		Name: "shop",
	}

	o.NoError(repository.Persists(o.ctx, shop))
	o.NoError(repository.Persists(o.ctx, shop))

	o.NoError(orm.Session(o.ctx).Flush())

	o.Equal(1, o.dbAdapter.InsertCounter())
}
