package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type PersistsTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	ctx       context.Context
}

func (o *PersistsTS) SetupSuite() {
	o.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	err := createSchema(o.pgDb)

	o.dbAdapter = helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.d3Orm = orm.NewOrm(o.dbAdapter)
	o.NoError(o.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	o.NoError(err)
}

func (o *PersistsTS) SetupTest() {
	o.ctx = o.d3Orm.CtxWithSession(context.Background())
}

func (o *PersistsTS) TearDownSuite() {
	o.NoError(deleteSchema(o.pgDb))
}

func (o *PersistsTS) TearDownTest() {
	o.dbAdapter.ResetCounters()
	o.NoError(clearSchema(o.pgDb))
}

func TestPersistsSuite(t *testing.T) {
	suite.Run(t, new(PersistsTS))
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
	o.NoError(orm.SessionFromCtx(o.ctx).Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='simple-shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is simple tests shop'")
}

func (o *PersistsTS) TestBigInsert() {
	shop, err := createAndPersistsShop(o.ctx, o.d3Orm)
	o.NoError(err)

	o.NoError(orm.SessionFromCtx(o.ctx).Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Authors.Get(0).(*Author).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Authors.Get(0).(*Author).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is tests shop'").
		SeeTwo("SELECT * FROM book_p").
		SeeThree("SELECT * FROM author_p").
		SeeFour("SELECT * FROM book_author_p")
}

func (o *PersistsTS) TestNoNewQueriesIfDoubleFlush() {
	_, err := createAndPersistsShop(o.ctx, o.d3Orm)
	o.NoError(err)

	o.NoError(orm.SessionFromCtx(o.ctx).Flush())
	insertCounter, updCounter := o.dbAdapter.InsertCounter(), o.dbAdapter.UpdateCounter()

	o.NoError(orm.SessionFromCtx(o.ctx).Flush())

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

	o.NoError(orm.SessionFromCtx(o.ctx).Flush())

	o.Equal(1, o.dbAdapter.InsertCounter())
}
