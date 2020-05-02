package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"d3/test/helpers"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PersistsTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	session   *orm.Session
}

func (o *PersistsTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	err := createSchema(o.pgDb)

	o.dbAdapter = helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.d3Orm = orm.NewOrm(o.dbAdapter)
	o.NoError(o.d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	o.NoError(err)
}

func (o *PersistsTS) SetupTest() {
	o.session = o.d3Orm.CreateSession()
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
	repository, err := o.d3Orm.CreateRepository(o.session, (*Shop)(nil))
	o.NoError(err)

	shop := &Shop{
		Books: nil,
		Profile: entity.NewWrapEntity(&ShopProfile{
			Description: "this is simple test shop",
		}),
		Name: "simple-shop",
	}

	o.NoError(repository.Persists(shop))
	o.NoError(o.session.Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='simple-shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is simple test shop'")
}

func (o *PersistsTS) TestBigInsert() {
	shop, err := createAndPersistsShop(o.d3Orm, o.session)
	o.NoError(err)

	o.NoError(o.session.Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Id.Int32)
	o.NotEqual(0, shop.Books.Get(0).(*Book).Authors.Get(0).(*Author).Id.Int32)
	o.NotEqual(0, shop.Books.Get(1).(*Book).Authors.Get(0).(*Author).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is test shop'").
		SeeTwo("SELECT * FROM book_p").
		SeeThree("SELECT * FROM author_p").
		SeeFour("SELECT * FROM book_author_p")
}

func (o *PersistsTS) TestNoNewQueriesIfDoubleFlush() {
	session := o.d3Orm.CreateSession()

	_, err := createAndPersistsShop(o.d3Orm, session)
	o.NoError(err)

	o.NoError(session.Flush())
	insertCounter, updCounter := o.dbAdapter.InsertCounter(), o.dbAdapter.UpdateCounter()

	o.NoError(session.Flush())

	o.Equal(insertCounter, o.dbAdapter.InsertCounter())
	o.Equal(updCounter, o.dbAdapter.UpdateCounter())
}

func (o *PersistsTS) TestOneNewEntityIfDoublePersist() {
	session := o.d3Orm.CreateSession()

	repository, _ := o.d3Orm.CreateRepository(session, (*Shop)(nil))

	shop := &Shop{
		Name: "shop",
	}

	o.NoError(repository.Persists(shop))
	o.NoError(repository.Persists(shop))

	o.NoError(session.Flush())

	o.Equal(1, o.dbAdapter.InsertCounter())
}
