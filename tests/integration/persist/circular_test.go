package persist

import (
	"context"
	"github.com/godzie44/d3/adapter"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type PersistsCircularTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	orm       *orm.Orm
}

func (o *PersistsCircularTS) SetupSuite() {
	o.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop_c(
		id SERIAL PRIMARY KEY,
		profile_id integer,
		friend_id integer, --for tests circular ref
		top_seller_id integer,
		name character varying(200) NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS profile_c(
		id SERIAL PRIMARY KEY,
		shop_id integer, --for tests circular ref
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

	o.dbAdapter = helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	o.orm = orm.NewOrm(o.dbAdapter)
	o.Assert().NoError(o.orm.Register(
		(*ShopCirc)(nil),
		(*ShopProfileCirc)(nil),
		(*SellerCirc)(nil),
	))
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
	o.dbAdapter.ResetCounters()
}

func TestPersistsCircularSuite(t *testing.T) {
	suite.Run(t, new(PersistsCircularTS))
}

func (o *PersistsCircularTS) TestInsertWithCircularRef() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*ShopCirc)(nil))

	profile := &ShopProfileCirc{
		Description: "shop profile",
	}
	shop := &ShopCirc{
		Profile: entity.NewCell(profile),
		Name:    "shop",
	}
	profile.Shop = entity.NewCell(shop)

	o.Assert().NoError(repository.Persists(ctx, shop))
	o.Assert().NoError(orm.Session(ctx).Flush())

	o.Assert().NotEqual(0, shop.Id.Int32)
	o.Assert().NotEqual(0, shop.Profile.Unwrap().(*ShopProfileCirc).Id.Int32)

	o.Assert().Equal(2, o.dbAdapter.InsertCounter())
	o.Assert().Equal(1, o.dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeOne("SELECT * FROM shop_c WHERE name='shop' AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_c WHERE description='shop profile' AND shop_id IS NOT NULL")
}

func (o *PersistsCircularTS) TestInsertWithSelfCircularRef() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*ShopCirc)(nil))

	shop1 := &ShopCirc{Name: "shop1"}
	shop2 := &ShopCirc{Name: "shop2"}

	shop3 := &ShopCirc{Name: "shop3"}
	shop4 := &ShopCirc{Name: "shop4"}
	shop5 := &ShopCirc{Name: "shop5"}

	shop1.FriendShop = entity.NewCell(shop2)
	shop2.FriendShop = entity.NewCell(shop1)

	shop3.FriendShop = entity.NewCell(shop4)
	shop4.FriendShop = entity.NewCell(shop5)
	shop5.FriendShop = entity.NewCell(shop3)

	o.Assert().NoError(repository.Persists(ctx, shop2))
	o.Assert().NoError(repository.Persists(ctx, shop3))

	o.Assert().NoError(orm.Session(ctx).Flush())

	o.Assert().NotEqual(0, shop1.Id.Int32)
	o.Assert().NotEqual(0, shop2.Id.Int32)
	o.Assert().NotEqual(0, shop3.Id.Int32)
	o.Assert().NotEqual(0, shop4.Id.Int32)
	o.Assert().NotEqual(0, shop5.Id.Int32)

	o.Assert().Equal(5, o.dbAdapter.InsertCounter())
	o.Assert().Equal(2, o.dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeTwo("SELECT * FROM shop_c WHERE name in ('shop1', 'shop2') AND friend_id IS NOT NULL").
		SeeThree("SELECT * FROM shop_c WHERE name in ('shop3', 'shop4', 'shop5') AND friend_id IS NOT NULL")
}

func (o *PersistsCircularTS) TestBigCircularReferenceGraph() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*ShopCirc)(nil))

	shop1 := &ShopCirc{Name: "shop1"}
	shop2 := &ShopCirc{Name: "shop2"}

	seller1 := &SellerCirc{Name: "Ivan"}
	seller2 := &SellerCirc{Name: "Sergej"}
	seller3 := &SellerCirc{Name: "Nickolay"}

	shop1.Sellers = entity.NewCollection(seller1)
	shop2.Sellers = entity.NewCollection(seller2, seller3)
	shop1.KnownSellers = entity.NewCollection(seller1, seller2)
	shop2.KnownSellers = entity.NewCollection(seller2, seller3)
	shop1.TopSeller = entity.NewCell(seller1)
	shop2.TopSeller = entity.NewCell(seller2)

	seller1.CurrentShop = entity.NewCell(shop1)
	seller1.KnownShops = entity.NewCollection(shop1)
	seller2.CurrentShop = entity.NewCell(shop2)
	seller2.KnownShops = entity.NewCollection(shop1, shop2)
	seller3.CurrentShop = entity.NewCell(shop2)
	seller3.KnownShops = entity.NewCollection(shop2)

	o.Assert().NoError(repository.Persists(ctx, shop1))
	o.Assert().NoError(repository.Persists(ctx, shop2))

	o.Assert().NoError(orm.Session(ctx).Flush())

	o.Assert().Equal(9, o.dbAdapter.InsertCounter())
	o.Assert().Equal(4, o.dbAdapter.UpdateCounter())

	helpers.NewPgTester(o.T(), o.pgDb).
		SeeTwo("SELECT * FROM shop_c WHERE name in ('shop1', 'shop2') AND top_seller_id IS NOT NULL").
		SeeThree("SELECT * FROM seller_c WHERE name in ('Ivan', 'Sergej', 'Nickolay') AND shop_id IS NOT NULL").
		SeeFour("SELECT * FROM known_shop_seller_c")
}
