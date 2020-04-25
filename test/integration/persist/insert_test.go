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
	pgDb *pgx.Conn
}

func (o *PersistsTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	err := createSchema(o.pgDb)
	o.NoError(err)
}

func (o *PersistsTS) TearDownSuite() {
	o.NoError(deleteSchema(o.pgDb))
}

func (o *PersistsTS) TearDownTest() {
	o.NoError(clearSchema(o.pgDb))
}

func TestPersistsSuite(t *testing.T) {
	suite.Run(t, new(PersistsTS))
}

func (o *PersistsTS) TestSimpleInsert() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	o.NoError(err)

	shop := &Shop{
		Books: nil,
		Profile: entity.NewWrapEntity(&ShopProfile{
			Description: "this is simple test shop",
		}),
		Name: "simple-shop",
	}

	o.NoError(repository.Persists(shop))
	o.NoError(session.Flush())

	o.NotEqual(0, shop.Id.Int32)
	o.NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='simple-shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is simple test shop'")
}

func (o *PersistsTS) TestBigInsert() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()
	shop, err := createAndPersistsShop(d3Orm, session)
	o.NoError(err)

	o.NoError(session.Flush())

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
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	o.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	_, err := createAndPersistsShop(d3Orm, session)
	o.NoError(err)

	o.NoError(session.Flush())
	insertCounter, updCounter := dbAdapter.InsertCounter(), dbAdapter.UpdateCounter()

	o.NoError(session.Flush())

	o.Equal(insertCounter, dbAdapter.InsertCounter())
	o.Equal(updCounter, dbAdapter.UpdateCounter())
}
