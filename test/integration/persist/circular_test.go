package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"d3/test/helpers"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PersistsCircularTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *PersistsCircularTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop_c(
		id SERIAL PRIMARY KEY,
		profile_id integer,
		friend_id integer, --for test circular ref
		top_seller_id integer,
		name character varying(200) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS profile_c(
		id SERIAL PRIMARY KEY,
		shop_id integer, --for test circular ref
		description character varying(1000) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS seller_c(
		id SERIAL PRIMARY KEY,
		name text NOT NULL,
		shop_id integer
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS known_shop_seller_c(
		shop_id integer NOT NULL,
		seller_id integer NOT NULL
	)`)
	o.Assert().NoError(err)
}

func (o *PersistsCircularTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE shop_c;
DROP TABLE profile_c;
DROP TABLE seller_c;
DROP TABLE known_shop_seller_c;
`)
	o.Assert().NoError(err)
}

func (o *PersistsCircularTS) TearDownTest() {
	_, err := o.pgDb.Exec(context.Background(), `
delete from shop_c;
delete from profile_c;
delete from seller_c;
delete from known_shop_seller_c;
`)
	o.Assert().NoError(err)
}

func TestPersistsCircularSuite(t *testing.T) {
	suite.Run(t, new(PersistsCircularTS))
}

type ShopCirc struct {
	entity struct{}      `d3:"table_name:shop_c"`
	Id     sql.NullInt32 `d3:"pk:auto"`
	Name   string

	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/ShopProfileCirc,join_on:profile_id>,type:lazy"`

	FriendShop entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/ShopCirc,join_on:friend_id>,type:lazy"`

	TopSeller    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/SellerCirc,join_on:top_seller_id>,type:lazy"`
	Sellers      entity.Collection    `d3:"one_to_many:<target_entity:d3/test/integration/persist/SellerCirc,join_on:shop_id>,type:lazy"`
	KnownSellers entity.Collection    `d3:"many_to_many:<target_entity:d3/test/integration/persist/SellerCirc,join_on:shop_id,reference_on:seller_id,join_table:known_shop_seller_c>,type:lazy"`
}

type ShopProfileCirc struct {
	entity      struct{}             `d3:"table_name:profile_c"`
	Id          sql.NullInt32        `d3:"pk:auto"`
	Shop        entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/ShopCirc,join_on:shop_id>,type:lazy"`
	Description string
}

type SellerCirc struct {
	entity struct{}      `d3:"table_name:seller_c"`
	Id     sql.NullInt32 `d3:"pk:auto"`
	Name   string

	CurrentShop entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/persist/ShopCirc,join_on:shop_id>,type:lazy"`
	KnownShops  entity.Collection    `d3:"many_to_many:<target_entity:d3/test/integration/persist/ShopCirc,join_on:seller_id,reference_on:shop_id,join_table:known_shop_seller_c>,type:lazy"`
}

func (o *PersistsCircularTS) TestInsertWithCircularRef() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	o.Assert().NoError(d3Orm.Register((*ShopCirc)(nil), (*ShopProfileCirc)(nil), (*SellerCirc)(nil)))

	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*ShopCirc)(nil))

	profile := &ShopProfileCirc{
		Description: "shop profile",
	}
	shop := &ShopCirc{
		Profile: entity.NewWrapEntity(profile),
		Name:    "shop",
	}
	profile.Shop = entity.NewWrapEntity(shop)

	o.Assert().NoError(repository.Persists(shop))
	o.Assert().NoError(session.Flush())

	o.Assert().NotEqual(0, shop.Id.Int32)
	o.Assert().NotEqual(0, shop.Profile.Unwrap().(*ShopProfileCirc).Id.Int32)

	o.Assert().Equal(2, dbAdapter.InsertCounter())
	o.Assert().Equal(1, dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_c WHERE name='shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_c WHERE description='shop profile' AND shop_id IS NOT NULL")
}

func (o *PersistsCircularTS) TestInsertWithSelfCircularRef() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	_ = d3Orm.Register((*ShopCirc)(nil), (*ShopProfileCirc)(nil), (*SellerCirc)(nil))

	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*ShopCirc)(nil))

	shop1 := &ShopCirc{Name: "shop1"}
	shop2 := &ShopCirc{Name: "shop2"}

	shop3 := &ShopCirc{Name: "shop3"}
	shop4 := &ShopCirc{Name: "shop4"}
	shop5 := &ShopCirc{Name: "shop5"}

	shop1.FriendShop = entity.NewWrapEntity(shop2)
	shop2.FriendShop = entity.NewWrapEntity(shop1)

	shop3.FriendShop = entity.NewWrapEntity(shop4)
	shop4.FriendShop = entity.NewWrapEntity(shop5)
	shop5.FriendShop = entity.NewWrapEntity(shop3)

	o.Assert().NoError(repository.Persists(shop2))
	o.Assert().NoError(repository.Persists(shop3))

	o.Assert().NoError(session.Flush())

	o.Assert().NotEqual(0, shop1.Id.Int32)
	o.Assert().NotEqual(0, shop2.Id.Int32)
	o.Assert().NotEqual(0, shop3.Id.Int32)
	o.Assert().NotEqual(0, shop4.Id.Int32)
	o.Assert().NotEqual(0, shop5.Id.Int32)

	o.Assert().Equal(5, dbAdapter.InsertCounter())
	o.Assert().Equal(2, dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeTwo("SELECT * FROM shop_c WHERE name in ('shop1', 'shop2') AND friend_id IS NOT NULL").
		SeeThree("SELECT * FROM shop_c WHERE name in ('shop3', 'shop4', 'shop5') AND friend_id IS NOT NULL")
}

func (o *PersistsCircularTS) TestBigCircularReferenceGraph() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	_ = d3Orm.Register((*ShopCirc)(nil), (*ShopProfileCirc)(nil), (*SellerCirc)(nil))

	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*ShopCirc)(nil))

	shop1 := &ShopCirc{Name: "shop1"}
	shop2 := &ShopCirc{Name: "shop2"}

	seller1 := &SellerCirc{Name: "Ivan"}
	seller2 := &SellerCirc{Name: "Sergej"}
	seller3 := &SellerCirc{Name: "Nickolay"}

	shop1.Sellers = entity.NewCollection([]interface{}{seller1})
	shop2.Sellers = entity.NewCollection([]interface{}{seller2, seller3})
	shop1.KnownSellers = entity.NewCollection([]interface{}{seller1, seller2})
	shop2.KnownSellers = entity.NewCollection([]interface{}{seller2, seller3})
	shop1.TopSeller = entity.NewWrapEntity(seller1)
	shop2.TopSeller = entity.NewWrapEntity(seller2)

	seller1.CurrentShop = entity.NewWrapEntity(shop1)
	seller1.KnownShops = entity.NewCollection([]interface{}{shop1})
	seller2.CurrentShop = entity.NewWrapEntity(shop2)
	seller2.KnownShops = entity.NewCollection([]interface{}{shop1, shop2})
	seller3.CurrentShop = entity.NewWrapEntity(shop2)
	seller3.KnownShops = entity.NewCollection([]interface{}{shop2})

	o.Assert().NoError(repository.Persists(shop1))
	o.Assert().NoError(repository.Persists(shop2))

	o.Assert().NoError(session.Flush())

	o.Assert().Equal(9, dbAdapter.InsertCounter())
	o.Assert().Equal(4, dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeTwo("SELECT * FROM shop_c WHERE name in ('shop1', 'shop2') AND top_seller_id IS NOT NULL").
		SeeThree("SELECT * FROM seller_c WHERE name in ('Ivan', 'Sergej', 'Nickolay') AND shop_id IS NOT NULL").
		SeeFour("SELECT * FROM known_shop_seller_c")
}
