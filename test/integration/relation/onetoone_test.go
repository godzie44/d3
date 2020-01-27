package relation

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
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

	_, err := o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_t1(
		id integer NOT NULL,
		data text NOT NULL,
		t2_id integer,
		CONSTRAINT test_entity_t1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_t2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t3_id integer,
		CONSTRAINT test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_t3(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT test_entity_t3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`
INSERT INTO test_entity_t1(id, data, t2_id) VALUES (1, 'entity_1_data', 1);
INSERT INTO test_entity_t1(id, data) VALUES (2, 'entity_1_data_2');
INSERT INTO test_entity_t2(id, data, t3_id) VALUES (1, 'entity_2_data', 1);
INSERT INTO test_entity_t3(id, data) VALUES (1, 'entity_3_data');
`)
	o.Assert().NoError(err)

}

func (o *OneToOneRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(),`
DROP TABLE test_entity_t1;
DROP TABLE test_entity_t2;
DROP TABLE test_entity_t3;
`)
	o.Assert().NoError(err)
}

func TestOneToOneRunTestSuite(t *testing.T) {
	suite.Run(t, new(OneToOneRelationTS))
}

type testEntity1OneToOneLL struct {
	entity struct{}             `d3:"table_name:test_entity_t1"`
	Id     int32                `d3:"pk:auto"`
	Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/testEntity2OneToOneLL,join_on:t2_id>,type:lazy"`
	Data   string
}

type testEntity2OneToOneLL struct {
	entity struct{}             `d3:"table_name:test_entity_t2"`
	Id     int32                `d3:"pk:auto"`
	Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/testEntity3OneToOneLL,join_on:t3_id,reference_on:id>,type:eager"`
	Data   string
}

type testEntity3OneToOneLL struct {
	entity struct{} `d3:"table_name:test_entity_t3"`
	Id     int32    `d3:"pk:auto"`
	Data   string
}

func (o *OneToOneRelationTS)TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))

	err := d3Orm.Register((*testEntity1OneToOneLL)(nil), (*testEntity2OneToOneLL)(nil), (*testEntity3OneToOneLL)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*testEntity1OneToOneLL)(nil))
	o.Assert().NoError( err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("id = ?", 1))
	o.Assert().NoError( err)

	o.Assert().IsType( &testEntity1OneToOneLL{}, entity)
	o.Assert().Equal( int32(1), entity.(*testEntity1OneToOneLL).Id)
	o.Assert().Equal( "entity_1_data", entity.(*testEntity1OneToOneLL).Data)

	relatedEntity := entity.(*testEntity1OneToOneLL).Rel.Unwrap().(*testEntity2OneToOneLL)
	o.Assert().IsType( &testEntity2OneToOneLL{}, relatedEntity)
	o.Assert().Equal( int32(1), relatedEntity.Id)
	o.Assert().Equal( "entity_2_data", relatedEntity.Data)
}

type testEntity1OneToOneEL struct {
	entity struct{}             `d3:"table_name:test_entity_t1"`
	Id     int32                `d3:"pk:auto"`
	Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/testEntity2OneToOneLL,join_on:t2_id,reference_on:id>,type:eager"`
	Data   string
}

func (o *OneToOneRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	_ = d3Orm.Register((*testEntity1OneToOneEL)(nil), (*testEntity2OneToOneLL)(nil), (*testEntity3OneToOneLL)(nil))

	session := d3Orm.CreateSession()
	repository, _ := d3Orm.CreateRepository(session, (*testEntity1OneToOneEL)(nil))
	e, err := repository.FindOne(repository.CreateQuery().AndWhere("test_entity_t1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType( &testEntity1OneToOneEL{}, e)
	o.Assert().Equal( int32(1), e.(*testEntity1OneToOneEL).Id)
	o.Assert().Equal( "entity_1_data", e.(*testEntity1OneToOneEL).Data)

	relatedEntity2 := e.(*testEntity1OneToOneEL).Rel.Unwrap().(*testEntity2OneToOneLL)
	o.Assert().IsType( &testEntity2OneToOneLL{}, relatedEntity2)
	o.Assert().Equal( int32(1), relatedEntity2.Id)
	o.Assert().Equal( "entity_2_data", relatedEntity2.Data)

	relatedEntity3 := relatedEntity2.Rel.Unwrap().(*testEntity3OneToOneLL)
	o.Assert().IsType( &testEntity3OneToOneLL{}, relatedEntity3)
	o.Assert().Equal( int32(1), relatedEntity3.Id)
	o.Assert().Equal( "entity_3_data", relatedEntity3.Data)
}

func (o *OneToOneRelationTS)TestEagerRelationNoRelated() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	_ = d3Orm.Register((*testEntity1OneToOneEL)(nil), (*testEntity2OneToOneLL)(nil), (*testEntity3OneToOneLL)(nil))

	session := d3Orm.CreateSession()
	repository, _ := d3Orm.CreateRepository(session, (*testEntity1OneToOneEL)(nil))
	e, _ := repository.FindOne(repository.CreateQuery().AndWhere("test_entity_t1.id = ?", 2))

	o.Assert().IsType( &testEntity1OneToOneEL{}, e)
	o.Assert().Equal( int32(2), e.(*testEntity1OneToOneEL).Id)
	o.Assert().Equal( "entity_1_data_2", e.(*testEntity1OneToOneEL).Data)

	o.Assert().True(e.(*testEntity1OneToOneEL).Rel.IsNil())
}

//func (o *OneToOneRelationTS)TestOneToOneEagerRelationNoRelated2() {
//	stormOrm := orm.NewOrm(adapter.NewGoPgAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
//	_ = stormOrm.Register((*testEntity1OneToOneEL)(nil), (*testEntity2OneToOneLL)(nil), (*testEntity3OneToOneLL)(nil))
//
//	session := stormOrm.CreateSession()
//	repository, _ := stormOrm.CreateRepository(session, (*testEntity1OneToOneEL)(nil))
//	e, _ := repository.FindAll(query.NewQuery())
//
//	o.Assert().IsType( []*testEntity1OneToOneEL{}, e)
//}
