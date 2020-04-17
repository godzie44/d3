package relation

import (
	"context"
	"d3/adapter"
	"d3/mapper"
	"d3/orm"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OneToManyRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *OneToManyRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

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

}

func (o *OneToManyRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE shop;
DROP TABLE book;
DROP TABLE discount;
`)
	o.Assert().NoError(err)
}

type ShopLR struct {
	entity struct{}          `d3:"table_name:shop"`
	Id     int32             `d3:"pk:auto"`
	Books  mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/BookLR,join_on:t1_id>,type:lazy"`
	Name   string
}

type BookLR struct {
	entity struct{} `d3:"table_name:book"`
	Id     int32    `d3:"pk:auto"`
	//Profile    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/PhotoLL,join_on:t3_id,reference_on:id>,type:eager"`
	Name string
}

func TestRunOneToManyTestSuite(t *testing.T) {
	suite.Run(t, new(OneToManyRelationTS))
}

func (o *OneToManyRelationTS) TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*ShopLR)(nil), (*BookLR)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*ShopLR)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("id = ?", 1))
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

type ShopER struct {
	entity struct{}          `d3:"table_name:shop"`
	Id     int32             `d3:"pk:auto"`
	Books  mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/BookER,join_on:t1_id,reference_on:id>,type:eager"`
	Name   string
}

type BookER struct {
	entity    struct{}          `d3:"table_name:book"`
	Id        int32             `d3:"pk:auto"`
	Discounts mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/DiscountER,join_on:t2_id,reference_on:id>,type:eager"`
	Name      string
}

type DiscountER struct {
	entity struct{} `d3:"table_name:discount"`
	Id     int32    `d3:"pk:auto"`
	Value  int32
}

func (o *OneToManyRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*ShopER)(nil), (*BookER)(nil), (*DiscountER)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*ShopER)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("shop.id = ?", 1))
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
