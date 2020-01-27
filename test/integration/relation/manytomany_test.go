package relation

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"d3/adapter"
	"d3/mapper"
	"d3/orm"
	"testing"
)

type ManyToManyRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *ManyToManyRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_1(
		id integer NOT NULL,
		data text NOT NULL,
		CONSTRAINT test_entity_1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT test_entity_2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS t1_t2(
		t1_id integer NOT NULL,
		t2_id integer NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS test_entity_3(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT test_entity_3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`CREATE TABLE IF NOT EXISTS t2_t3(
		t2_id integer NOT NULL,
		t3_id integer NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(),`
INSERT INTO test_entity_1(id, data) VALUES (1, 'entity_1_data_1');
INSERT INTO test_entity_1(id, data) VALUES (2, 'entity_1_data_2');
INSERT INTO test_entity_1(id, data) VALUES (3, 'entity_1_data_3');
INSERT INTO test_entity_2(id, data) VALUES (1, 'entity_2_data_1');
INSERT INTO test_entity_2(id, data) VALUES (2, 'entity_2_data_2');
INSERT INTO test_entity_3(id, data) VALUES (1, 'entity_3_data_1');
INSERT INTO t1_t2(t1_id, t2_id) VALUES (1, 1);
INSERT INTO t1_t2(t1_id, t2_id) VALUES (1, 2);
INSERT INTO t1_t2(t1_id, t2_id) VALUES (2, 2);
INSERT INTO t1_t2(t1_id, t2_id) VALUES (3, 1);
INSERT INTO t2_t3(t2_id, t3_id) VALUES (1, 1);
`)
	o.Assert().NoError(err)

}

func (o *ManyToManyRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(),`
DROP TABLE test_entity_1;
DROP TABLE test_entity_2;
DROP TABLE test_entity_3;
DROP TABLE t1_t2;
DROP TABLE t2_t3;
`)
	o.Assert().NoError(err)
}

func TestManyToManyTestSuite(t *testing.T) {
	suite.Run(t, new(ManyToManyRelationTS))
}

type testEntity1ManyToManyLL struct {
	entity struct{}          `d3:"table_name:test_entity_1"`
	Id     int32             `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/testEntity2ManyToManyLL,join_on:t1_id,reference_on:t2_id,join_table:t1_t2>,type:lazy"`
	Data   string
}

type testEntity2ManyToManyLL struct {
	entity struct{} `d3:"table_name:test_entity_2"`
	Id     int32    `d3:"pk:auto"`
	Data   string
}

func (o *ManyToManyRelationTS) TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*testEntity1ManyToManyLL)(nil), (*testEntity2ManyToManyLL)(nil), (*testEntity3ManyToMany)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*testEntity1ManyToManyLL)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&testEntity1ManyToManyLL{}, entity)
	o.Assert().Equal(int32(1), entity.(*testEntity1ManyToManyLL).Id)
	o.Assert().Equal("entity_1_data_1", entity.(*testEntity1ManyToManyLL).Data)

	relatedEntities := entity.(*testEntity1ManyToManyLL).Rel
	o.Assert().Equal(relatedEntities.Count(), 2)
	o.Assert().Subset(
		[]string{"entity_2_data_1", "entity_2_data_2"},
		[]string{relatedEntities.Get(0).(*testEntity2ManyToManyLL).Data, relatedEntities.Get(1).(*testEntity2ManyToManyLL).Data},
	)
}

type testEntity1ManyToManyEL struct {
	entity struct{}          `d3:"table_name:test_entity_1"`
	Id     int32             `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/testEntity2ManyToManyEL,join_on:t1_id,reference_on:t2_id,join_table:t1_t2>,type:eager"`
	Data   string
}

type testEntity2ManyToManyEL struct {
	entity struct{}          `d3:"table_name:test_entity_2"`
	Id     int32             `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/testEntity3ManyToMany,join_on:t2_id,reference_on:t3_id,join_table:t2_t3>,type:eager"`
	Data   string
}

type testEntity3ManyToMany struct {
	entity struct{} `d3:"table_name:test_entity_3"`
	Id     int32    `d3:"pk:auto"`
	Data   string
}

func (o *ManyToManyRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*testEntity1ManyToManyEL)(nil), (*testEntity2ManyToManyEL)(nil), (*testEntity3ManyToMany)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*testEntity1ManyToManyEL)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&testEntity1ManyToManyEL{}, entity)
	o.Assert().Equal(int32(1), entity.(*testEntity1ManyToManyEL).Id)
	o.Assert().Equal("entity_1_data_1", entity.(*testEntity1ManyToManyEL).Data)

	relatedEntities := entity.(*testEntity1ManyToManyEL).Rel
	o.Assert().Equal(2, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"entity_2_data_1", "entity_2_data_2"},
		[]string{relatedEntities.Get(0).(*testEntity2ManyToManyEL).Data, relatedEntities.Get(1).(*testEntity2ManyToManyEL).Data},
	)

	if relatedEntities.Get(0).(*testEntity2ManyToManyEL).Rel.Count() != 1 &&  relatedEntities.Get(1).(*testEntity2ManyToManyEL).Rel.Count() != 1 {
		o.Assert().Fail("testEntity3 not found")
	}
}
