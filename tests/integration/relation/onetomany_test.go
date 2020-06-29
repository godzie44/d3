package relation

import (
	"context"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type OneToManyRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
	orm  *orm.Orm
}

func (o *OneToManyRelationTS) SetupSuite() {
	o.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop(
		id integer NOT NULL,
		name text NOT NULL,
		CONSTRAINT test_entity_t1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book(
		id integer NOT NULL,
		name character varying(200) NOT NULL,
		t1_id integer,
		CONSTRAINT test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS discount(
		id integer NOT NULL,
		value integer NOT NULL,
		t2_id integer,
		CONSTRAINT test_entity_t3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO shop(id, name) VALUES (1, 'book-shop');
INSERT INTO book(id, name, t1_id) VALUES (1, 'Antic Hay', 1);
INSERT INTO book(id, name, t1_id) VALUES (2, 'An Evil Cradling', 1);
INSERT INTO book(id, name, t1_id) VALUES (3, 'Arms and the Man', 1);
INSERT INTO discount(id, value, t2_id) VALUES (1, 33, 1);
`)
	o.Assert().NoError(err)

	o.orm = orm.New(d3pgx.NewPgxDriver(o.pgDb))
	o.NoError(o.orm.Register(
		(*ShopLR)(nil),
		(*BookLR)(nil),
		(*ShopER)(nil),
		(*BookER)(nil),
		(*DiscountER)(nil),
	))
}

func (o *OneToManyRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE shop;
DROP TABLE book;
DROP TABLE discount;
`)
	o.Assert().NoError(err)
}

func TestRunOneToManyTestSuite(t *testing.T) {
	suite.Run(t, new(OneToManyRelationTS))
}

func (o *OneToManyRelationTS) TestLazyRelation() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, err := o.orm.MakeRepository((*ShopLR)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(ctx, repository.Select().AndWhere("id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&ShopLR{}, entity)
	o.Assert().Equal(int32(1), entity.(*ShopLR).Id)
	o.Assert().Equal("book-shop", entity.(*ShopLR).Name)

	relatedEntities := entity.(*ShopLR).Books
	o.Assert().Equal(relatedEntities.Count(), 3)
	o.Assert().Subset(
		[]string{"Antic Hay", "An Evil Cradling", "Arms and the Man"},
		[]string{relatedEntities.Get(0).(*BookLR).Name, relatedEntities.Get(1).(*BookLR).Name, relatedEntities.Get(2).(*BookLR).Name},
	)
}

func (o *OneToManyRelationTS) TestEagerRelation() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, err := o.orm.MakeRepository((*ShopER)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(ctx, repository.Select().AndWhere("shop.id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&ShopER{}, entity)
	o.Assert().Equal(int32(1), entity.(*ShopER).Id)
	o.Assert().Equal("book-shop", entity.(*ShopER).Name)

	relatedEntities := entity.(*ShopER).Books
	o.Assert().Equal(relatedEntities.Count(), 3)
	o.Assert().Subset(
		[]string{"Antic Hay", "An Evil Cradling", "Arms and the Man"},
		[]string{relatedEntities.Get(0).(*BookER).Name, relatedEntities.Get(1).(*BookER).Name, relatedEntities.Get(2).(*BookER).Name},
	)
}
