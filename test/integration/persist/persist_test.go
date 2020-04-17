package persist

import (
	"context"
	"d3/adapter"
	"d3/mapper"
	"d3/orm"
	"d3/orm/entity"
	"d3/test/helpers"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Shop struct {
	entity  struct{}             `d3:"table_name:shop_p"`
	Id      sql.NullInt32        `d3:"pk:auto"`
	Books   mapper.Collection    `d3:"one_to_many:<target_entity:d3/test/integration/persist/Book,join_on:shop_id>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/ShopProfile,join_on:profile_id>,type:lazy"`
	Name    string
}

type ShopProfile struct {
	entity      struct{}      `d3:"table_name:profile_p"`
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

type Book struct {
	entity  struct{}          `d3:"table_name:book_p"`
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors mapper.Collection `d3:"many_to_many:<target_entity:d3/test/integration/persist/Author,join_on:book_id,reference_on:author_id,join_table:book_author_p>,type:lazy"`
	Name    string
}

type Author struct {
	entity struct{}      `d3:"table_name:author_p"`
	Id     sql.NullInt32 `d3:"pk:auto"`
	Name   string
}

type PersistsTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *PersistsTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop_p(
		id SERIAL PRIMARY KEY,
		profile_id integer,
		name character varying(200) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS profile_p(
		id SERIAL PRIMARY KEY,
		shop_id integer, --for test circular ref
		description character varying(1000) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book_p(
		id SERIAL PRIMARY KEY,
		name text NOT NULL,
		shop_id integer
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS author_p(
		id SERIAL PRIMARY KEY,
		name character varying(200) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book_author_p(
		book_id integer NOT NULL,
		author_id integer NOT NULL
	)`)
	o.Assert().NoError(err)
}

func (o *PersistsTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`)
	o.Assert().NoError(err)
}

func (o *PersistsTS) TearDownTest() {
	_, err := o.pgDb.Exec(context.Background(), `
delete from book_p;
delete from author_p;
delete from book_author_p;
delete from shop_p;
delete from profile_p;
`)
	o.Assert().NoError(err)
}

func TestPersistsSuite(t *testing.T) {
	suite.Run(t, new(PersistsTS))
}

func (o *PersistsTS) TestSimpleInsert() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.Assert().NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	o.Assert().NoError(err)

	shop := &Shop{
		Books: nil,
		Profile: entity.NewWrapEntity(&ShopProfile{
			Description: "this is simple test shop",
		}),
		Name: "simple-shop",
	}

	o.Assert().NoError(repository.Persists(shop))
	o.Assert().NoError(session.Flush())

	o.Assert().NotEqual(0, shop.Id.Int32)
	o.Assert().NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='simple-shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='this is simple test shop'")
}

func (o *PersistsTS) TestBigInsert() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.Assert().NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()
	shop, err := createAndPersistsShop(d3Orm, session)
	o.Assert().NoError(err)

	o.Assert().NoError(session.Flush())

	o.Assert().NotEqual(0, shop.Id.Int32)
	o.Assert().NotEqual(0, shop.Profile.Unwrap().(*ShopProfile).Id.Int32)
	o.Assert().NotEqual(0, shop.Books.Get(0).(*Book).Id.Int32)
	o.Assert().NotEqual(0, shop.Books.Get(1).(*Book).Id.Int32)
	o.Assert().NotEqual(0, shop.Books.Get(0).(*Book).Authors.Get(0).(*Author).Id.Int32)
	o.Assert().NotEqual(0, shop.Books.Get(1).(*Book).Authors.Get(0).(*Author).Id.Int32)

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
	o.Assert().NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	_, err := createAndPersistsShop(d3Orm, session)
	o.Assert().NoError(err)

	o.Assert().NoError(session.Flush())
	insertCounter, updCounter := dbAdapter.InsertCounter(), dbAdapter.UpdateCounter()

	o.Assert().NoError(session.Flush())

	o.Assert().Equal(insertCounter, dbAdapter.InsertCounter())
	o.Assert().Equal(updCounter, dbAdapter.UpdateCounter())
}

func (o *PersistsTS) TestInsertThenUpdate() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	o.Assert().NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	o.Assert().NoError(err)

	o.Assert().NoError(session.Flush())

	shop.Name = "new shop"

	o.Assert().NoError(session.Flush())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='new shop'").
		See(0, "SELECT * FROM shop_p WHERE name='shop'")

	o.Assert().Equal(1, dbAdapter.UpdateCounter())
}

func (o *PersistsTS) TestInsertThenUpdateRelations() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	o.Assert().NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	o.Assert().NoError(err)

	o.Assert().NoError(session.Flush())

	shop.Books.Remove(0)

	o.Assert().NoError(session.Flush())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NOT NULL").
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NULL")

	o.Assert().Equal(1, dbAdapter.UpdateCounter())
}

func createAndPersistsShop(orm *orm.Orm, s *orm.Session) (*Shop, error) {
	repository, _ := orm.CreateRepository(s, (*Shop)(nil))

	author1 := &Author{
		Name: "author1",
	}
	shop := &Shop{
		Books: mapper.NewCollection([]interface{}{&Book{
			Authors: mapper.NewCollection([]interface{}{author1, &Author{Name: "author 2"}}),
			Name:    "book 1",
		}, &Book{
			Authors: mapper.NewCollection([]interface{}{author1, &Author{Name: "author 3"}}),
			Name:    "book 2",
		}}),
		Profile: entity.NewWrapEntity(&ShopProfile{
			Description: "this is test shop",
		}),
		Name: "shop",
	}

	return shop, repository.Persists(shop)
}
