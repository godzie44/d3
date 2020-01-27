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

type OneToManyRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *OneToManyRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_1(
		id integer NOT NULL,
		data text NOT NULL,
		CONSTRAINT test_entity_t1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t1_id integer,
		CONSTRAINT test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_3(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t2_id integer,
		CONSTRAINT test_entity_t3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO test_entity_1(id, data) VALUES (1, 'entity_1_data');
INSERT INTO test_entity_2(id, data, t1_id) VALUES (1, 'entity_2_data_1', 1);
INSERT INTO test_entity_2(id, data, t1_id) VALUES (2, 'entity_2_data_2', 1);
INSERT INTO test_entity_2(id, data, t1_id) VALUES (3, 'entity_2_data_3', 1);
INSERT INTO test_entity_3(id, data, t2_id) VALUES (1, 'entity_3_data', 1);
`)
	o.Assert().NoError(err)

}

func (o *OneToManyRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE test_entity_1;
DROP TABLE test_entity_2;
DROP TABLE test_entity_3;
`)
	o.Assert().NoError(err)
}

type testEntity1OneToManyLR struct {
	entity struct{}          `d3:"table_name:test_entity_1"`
	Id     int32             `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/testEntity2OneToManyLR,join_on:t1_id>,type:lazy"`
	Data   string
}

type testEntity2OneToManyLR struct {
	entity struct{} `d3:"table_name:test_entity_2"`
	Id     int32    `d3:"pk:auto"`
	//Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/testEntity3OneToOneLL,join_on:t3_id,reference_on:id>,type:eager"`
	Data string
}

func TestRunOneToManyTestSuite(t *testing.T) {
	suite.Run(t, new(OneToManyRelationTS))
}

func (o *OneToManyRelationTS) TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*testEntity1OneToManyLR)(nil), (*testEntity2OneToManyLR)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*testEntity1OneToManyLR)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&testEntity1OneToManyLR{}, entity)
	o.Assert().Equal(int32(1), entity.(*testEntity1OneToManyLR).Id)
	o.Assert().Equal("entity_1_data", entity.(*testEntity1OneToManyLR).Data)

	relatedEntities := entity.(*testEntity1OneToManyLR).Rel
	o.Assert().Equal(relatedEntities.Count(), 3)
	o.Assert().Subset(
		[]string{"entity_2_data_1", "entity_2_data_2", "entity_2_data_3"},
		[]string{relatedEntities.Get(0).(*testEntity2OneToManyLR).Data, relatedEntities.Get(1).(*testEntity2OneToManyLR).Data, relatedEntities.Get(2).(*testEntity2OneToManyLR).Data},
	)
}

type testEntity1OneToManyER struct {
	entity struct{}          `d3:"table_name:test_entity_1"`
	Id     int32             `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/testEntity2OneToManyER,join_on:t1_id,reference_on:id>,type:eager"`
	Data   string
}

type testEntity2OneToManyER struct {
	entity struct{} `d3:"table_name:test_entity_2"`
	Id     int32    `d3:"pk:auto"`
	Rel    mapper.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/testEntity3OneToManyER,join_on:t2_id,reference_on:id>,type:eager"`
	Data string
}

type testEntity3OneToManyER struct {
	entity struct{} `d3:"table_name:test_entity_3"`
	Id     int32    `d3:"pk:auto"`
	Data string
}

func (o *OneToManyRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*testEntity1OneToManyER)(nil), (*testEntity2OneToManyER)(nil), (*testEntity3OneToManyER)(nil))
	o.Assert().NoError(err)

	session := d3Orm.CreateSession()
	repository, err := d3Orm.CreateRepository(session, (*testEntity1OneToManyER)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&testEntity1OneToManyER{}, entity)
	o.Assert().Equal(int32(1), entity.(*testEntity1OneToManyER).Id)
	o.Assert().Equal("entity_1_data", entity.(*testEntity1OneToManyER).Data)

	relatedEntities := entity.(*testEntity1OneToManyER).Rel
	o.Assert().Equal(relatedEntities.Count(), 3)
	o.Assert().Subset(
		[]string{"entity_2_data_1", "entity_2_data_2", "entity_2_data_3"},
		[]string{relatedEntities.Get(0).(*testEntity2OneToManyER).Data, relatedEntities.Get(1).(*testEntity2OneToManyER).Data, relatedEntities.Get(2).(*testEntity2OneToManyER).Data},
	)
}
