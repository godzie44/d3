package relation

import (
	"context"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OneToOneRelationTS struct {
	suite.Suite
	execSqlFn func(sql string) error
	orm       *orm.Orm
}

func (o *OneToOneRelationTS) SetupSuite() {
	err := o.execSqlFn(`CREATE TABLE IF NOT EXISTS shop(
		id integer NOT NULL,
		data text NOT NULL,
		t2_id integer,
		CONSTRAINT shop_pkey PRIMARY KEY (id)
	)`)
	o.NoError(err)

	err = o.execSqlFn(`CREATE TABLE IF NOT EXISTS profile(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t3_id integer,
		CONSTRAINT profile_pkey PRIMARY KEY (id)
	)`)
	o.NoError(err)

	err = o.execSqlFn(`CREATE TABLE IF NOT EXISTS photo(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT photo_pkey PRIMARY KEY (id)
	)`)
	o.NoError(err)

	err = o.execSqlFn(`
INSERT INTO shop(id, data, t2_id) VALUES (1, 'entity_1_data', 1);
INSERT INTO shop(id, data) VALUES (2, 'entity_1_data_2');
INSERT INTO profile(id, data, t3_id) VALUES (1, 'entity_2_data', 1);
INSERT INTO photo(id, data) VALUES (1, 'entity_3_data');
`)
	o.NoError(err)

	o.NoError(o.orm.Register(
		(*ShopLL)(nil),
		(*ProfileLL)(nil),
		(*PhotoLL)(nil),
		(*ShopEL)(nil),
	))
}

func (o *OneToOneRelationTS) TearDownSuite() {
	o.Assert().NoError(o.execSqlFn(`
DROP TABLE shop;
DROP TABLE profile;
DROP TABLE photo;
`))
}

func TestPGOneToOneTestSuite(t *testing.T) {
	_, d3orm, execSqlFn, _ := db.CreatePGTestComponents(t)

	mtmTS := &OneToOneRelationTS{
		orm:       d3orm,
		execSqlFn: execSqlFn,
	}
	suite.Run(t, mtmTS)
}

func TestSQLiteOneToOneTestSuite(t *testing.T) {
	_, d3orm, execSqlFn, _ := db.CreateSQLiteTestComponents(t)

	mtmTS := &OneToOneRelationTS{
		orm:       d3orm,
		execSqlFn: execSqlFn,
	}

	suite.Run(t, mtmTS)
}

func (o *OneToOneRelationTS) TestLazyRelation() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, err := o.orm.MakeRepository((*ShopLL)(nil))
	o.Assert().NoError(err)

	shop, err := repository.FindOne(ctx, repository.Select().Where("id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&ShopLL{}, shop)
	o.Assert().Equal(int32(1), shop.(*ShopLL).ID.Int32)
	o.Assert().Equal("entity_1_data", shop.(*ShopLL).Data)

	relatedEntity := shop.(*ShopLL).Profile.Unwrap().(*ProfileLL)
	o.Assert().IsType(&ProfileLL{}, relatedEntity)
	o.Assert().Equal(int32(1), relatedEntity.ID)
	o.Assert().Equal("entity_2_data", relatedEntity.Data)
}

func (o *OneToOneRelationTS) TestEagerRelation() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*ShopEL)(nil))
	e, err := repository.FindOne(ctx, repository.Select().Where("shop.id", "=", 1))
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
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*ShopEL)(nil))
	e, _ := repository.FindOne(ctx, repository.Select().Where("shop.id", "=", 2))

	o.Assert().IsType(&ShopEL{}, e)
	o.Assert().Equal(int32(2), e.(*ShopEL).Id)
	o.Assert().Equal("entity_1_data_2", e.(*ShopEL).Data)

	o.Assert().True(e.(*ShopEL).Profile.IsNil())
}
