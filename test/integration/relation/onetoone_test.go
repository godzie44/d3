package relation

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OneToOneRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *OneToOneRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop(
		id integer NOT NULL,
		data text NOT NULL,
		t2_id integer,
		CONSTRAINT shop_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS profile(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t3_id integer,
		CONSTRAINT profile_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS photo(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT photo_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO shop(id, data, t2_id) VALUES (1, 'entity_1_data', 1);
INSERT INTO shop(id, data) VALUES (2, 'entity_1_data_2');
INSERT INTO profile(id, data, t3_id) VALUES (1, 'entity_2_data', 1);
INSERT INTO photo(id, data) VALUES (1, 'entity_3_data');
`)
	o.Assert().NoError(err)

}

func (o *OneToOneRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE shop;
DROP TABLE profile;
DROP TABLE photo;
`)
	o.Assert().NoError(err)
}

func TestOneToOneRunTestSuite(t *testing.T) {
	suite.Run(t, new(OneToOneRelationTS))
}

type ShopLL struct {
	entity  struct{}             `d3:"table_name:shop"`
	ID      sql.NullInt32        `d3:"pk:auto"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/ProfileLL,join_on:t2_id>,type:lazy"`
	Data    string
}

type ProfileLL struct {
	entity struct{}             `d3:"table_name:profile"`
	ID     int32                `d3:"pk:auto"`
	Photo  entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/PhotoLL,join_on:t3_id,reference_on:id>,type:eager"`
	Data   string
}

type PhotoLL struct {
	entity struct{} `d3:"table_name:photo"`
	ID     int32    `d3:"pk:auto"`
	Data   string
}

func (o *OneToOneRelationTS) TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))

	err := d3Orm.Register((*ShopLL)(nil), (*ProfileLL)(nil), (*PhotoLL)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*ShopLL)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&ShopLL{}, entity)
	o.Assert().Equal(int32(1), entity.(*ShopLL).ID.Int32)
	o.Assert().Equal("entity_1_data", entity.(*ShopLL).Data)

	relatedEntity := entity.(*ShopLL).Profile.Unwrap().(*ProfileLL)
	o.Assert().IsType(&ProfileLL{}, relatedEntity)
	o.Assert().Equal(int32(1), relatedEntity.ID)
	o.Assert().Equal("entity_2_data", relatedEntity.Data)
}

type ShopEL struct {
	entity  struct{}             `d3:"table_name:shop"`
	Id      int32                `d3:"pk:auto"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/ProfileLL,join_on:t2_id,reference_on:id>,type:eager"`
	Data    string
}

func (o *OneToOneRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	_ = d3Orm.Register((*ShopEL)(nil), (*ProfileLL)(nil), (*PhotoLL)(nil))

	session := d3Orm.CreateSession()
	repository, _ := d3Orm.CreateRepository(session, (*ShopEL)(nil))
	e, err := repository.FindOne(repository.CreateQuery().AndWhere("shop.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&ShopEL{}, e)
	o.Assert().Equal(int32(1), e.(*ShopEL).Id)
	o.Assert().Equal("entity_1_data", e.(*ShopEL).Data)

	relatedEntity2 := e.(*ShopEL).Profile.Unwrap().(*ProfileLL)
	o.Assert().IsType(&ProfileLL{}, relatedEntity2)
	o.Assert().Equal(int32(1), relatedEntity2.ID)
	o.Assert().Equal("entity_2_data", relatedEntity2.Data)

	relatedEntity3 := relatedEntity2.Photo.Unwrap().(*PhotoLL)
	o.Assert().IsType(&PhotoLL{}, relatedEntity3)
	o.Assert().Equal(int32(1), relatedEntity3.ID)
	o.Assert().Equal("entity_3_data", relatedEntity3.Data)
}

func (o *OneToOneRelationTS) TestEagerRelationNoRelated() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	_ = d3Orm.Register((*ShopEL)(nil), (*ProfileLL)(nil), (*PhotoLL)(nil))

	session := d3Orm.CreateSession()
	repository, _ := d3Orm.CreateRepository(session, (*ShopEL)(nil))
	e, _ := repository.FindOne(repository.CreateQuery().AndWhere("shop.id = ?", 2))

	o.Assert().IsType(&ShopEL{}, e)
	o.Assert().Equal(int32(2), e.(*ShopEL).Id)
	o.Assert().Equal("entity_1_data_2", e.(*ShopEL).Data)

	o.Assert().True(e.(*ShopEL).Profile.IsNil())
}

//func (o *OneToOneRelationTS)TestOneToOneEagerRelationNoRelated2() {
//	stormOrm := orm.NewOrm(adapter.NewGoPgAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
//	_ = stormOrm.Register((*ShopEL)(nil), (*ProfileLL)(nil), (*PhotoLL)(nil))
//
//	session := stormOrm.CreateSession()
//	repository, _ := stormOrm.CreateRepository(session, (*ShopEL)(nil))
//	e, _ := repository.FindAll(query.NewQuery())
//
//	o.Assert().IsType( []*ShopEL{}, e)
//}
