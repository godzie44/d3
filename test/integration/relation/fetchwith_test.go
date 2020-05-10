package relation

import (
	"context"
	"d3/adapter"
	"d3/orm"
	d3entity "d3/orm/entity"
	"d3/test/helpers"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type FetchWithRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *FetchWithRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_1(
		id integer NOT NULL,
		data text NOT NULL,
		e2_id integer,
		CONSTRAINT test_entity_1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_3(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		e2_id integer,

		CONSTRAINT test_entity_t3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_entity_4(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		CONSTRAINT test_entity_t4_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS t3_t4(
		t3_id integer NOT NULL,
		t4_id integer NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO test_entity_1(id, data, e2_id) VALUES (1, 'entity_1_data_1', 1);
INSERT INTO test_entity_1(id, data, e2_id) VALUES (2, 'entity_1_data_2', 2);
INSERT INTO test_entity_1(id, data, e2_id) VALUES (3, 'entity_1_data_3', null);
INSERT INTO test_entity_2(id, data) VALUES (1, 'entity_2_data_1');
INSERT INTO test_entity_2(id, data) VALUES (2, 'entity_2_data_2');
INSERT INTO test_entity_3(id, data, e2_id) VALUES (1, 'entity_3_data_1', 1);
INSERT INTO test_entity_3(id, data, e2_id) VALUES (2, 'entity_3_data_2', 1);
INSERT INTO test_entity_3(id, data, e2_id) VALUES (3, 'entity_3_data_3', 1);
INSERT INTO test_entity_4(id, data) VALUES (1, 'entity_4_data_2');
INSERT INTO test_entity_4(id, data) VALUES (2, 'entity_4_data_2');
INSERT INTO test_entity_4(id, data) VALUES (3, 'entity_4_data_2');
INSERT INTO t3_t4(t3_id, t4_id) VALUES (1, 1);
INSERT INTO t3_t4(t3_id, t4_id) VALUES (1, 2);
INSERT INTO t3_t4(t3_id, t4_id) VALUES (2, 2);
`)
	o.Assert().NoError(err)

}

func (o *FetchWithRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE test_entity_1;
DROP TABLE test_entity_2;
DROP TABLE test_entity_3;
DROP TABLE test_entity_4;
DROP TABLE t3_t4;
`)
	o.Assert().NoError(err)
}

func TestFetchWithRelationTestSuite(t *testing.T) {
	suite.Run(t, new(FetchWithRelationTS))
}

type fwTestEntity1 struct {
	entity struct{}               `d3:"table_name:test_entity_1"` //nolint:unused,structcheck
	Id     int32                  `d3:"pk:auto"`
	Rel    d3entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/relation/fwTestEntity2,join_on:e2_id,reference_on:id>,type:lazy"`
	Data   string
}

type fwTestEntity2 struct {
	entity struct{}            `d3:"table_name:test_entity_2"` //nolint:unused,structcheck
	Id     int32               `d3:"pk:auto"`
	Rel    d3entity.Collection `d3:"one_to_many:<target_entity:d3/test/integration/relation/fwTestEntity3,join_on:e2_id>,type:lazy"`
	Data   string
}

func (o *FetchWithRelationTS) TestFetchWithOneToOne() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(wrappedDbAdapter)

	_ = d3Orm.Register((*fwTestEntity1)(nil), (*fwTestEntity2)(nil), (*fwTestEntity3)(nil), (*fwTestEntity4)(nil))
	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*fwTestEntity1)(nil))
	q := repository.CreateQuery()
	_ = q.AndWhere("test_entity_1.id = ?", 1).With("d3/test/integration/relation/fwTestEntity2")
	entity, _ := repository.FindOne(q)

	o.Assert().IsType(&fwTestEntity1{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity1).Id)
	o.Assert().Equal("entity_1_data_1", entity.(*fwTestEntity1).Data)

	relatedEntity := entity.(*fwTestEntity1).Rel.Unwrap().(*fwTestEntity2)
	o.Assert().Subset(
		[]interface{}{int32(1), "entity_2_data_1"},
		[]interface{}{relatedEntity.Id, relatedEntity.Data},
	)

	o.Assert().Equal(1, wrappedDbAdapter.QueryCounter())
}

type fwTestEntity3 struct {
	entity struct{}            `d3:"table_name:test_entity_3"` //nolint:unused,structcheck
	Id     int32               `d3:"pk:auto"`
	Rel    d3entity.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/fwTestEntity4,join_on:t3_id,reference_on:t4_id,join_table:t3_t4>,type:lazy"`
	Data   string
}

func (o *FetchWithRelationTS) TestFetchWithOneToMany() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(wrappedDbAdapter)

	_ = d3Orm.Register((*fwTestEntity1)(nil), (*fwTestEntity2)(nil), (*fwTestEntity3)(nil), (*fwTestEntity4)(nil))
	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*fwTestEntity2)(nil))
	q := repository.CreateQuery()
	_ = q.AndWhere("test_entity_2.id = ?", 1).With("d3/test/integration/relation/fwTestEntity3")
	entity, _ := repository.FindOne(q)

	o.Assert().IsType(&fwTestEntity2{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity2).Id)
	o.Assert().Equal("entity_2_data_1", entity.(*fwTestEntity2).Data)

	relatedEntities := entity.(*fwTestEntity2).Rel
	o.Assert().Equal(3, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"entity_3_data_1", "entity_3_data_2", "entity_3_data_3"},
		[]string{relatedEntities.Get(0).(*fwTestEntity3).Data, relatedEntities.Get(1).(*fwTestEntity3).Data, relatedEntities.Get(2).(*fwTestEntity3).Data},
	)

	o.Assert().Equal(1, wrappedDbAdapter.QueryCounter())
}

type fwTestEntity4 struct {
	entity struct{} `d3:"table_name:test_entity_4"` //nolint:unused,structcheck
	Id     int32    `d3:"pk:auto"`
	Data   string
}

func (o *FetchWithRelationTS) TestFetchWithManyToMany() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(wrappedDbAdapter)

	_ = d3Orm.Register((*fwTestEntity1)(nil), (*fwTestEntity2)(nil), (*fwTestEntity3)(nil), (*fwTestEntity4)(nil))

	repository, _ := d3Orm.MakeSession().MakeRepository((*fwTestEntity3)(nil))

	q := repository.CreateQuery()
	_ = q.AndWhere("test_entity_3.id = ?", 1).With("d3/test/integration/relation/fwTestEntity4")
	entity, _ := repository.FindOne(q)

	o.Assert().IsType(&fwTestEntity3{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity3).Id)
	o.Assert().Equal("entity_3_data_1", entity.(*fwTestEntity3).Data)

	relatedEntities := entity.(*fwTestEntity3).Rel
	o.Assert().Equal(2, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"entity_4_data_1", "entity_4_data_2"},
		[]string{relatedEntities.Get(0).(*fwTestEntity4).Data, relatedEntities.Get(1).(*fwTestEntity4).Data},
	)

	o.Assert().Equal(1, wrappedDbAdapter.QueryCounter())
}

func (o *FetchWithRelationTS) TestFetchFullGraph() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(wrappedDbAdapter)

	_ = d3Orm.Register((*fwTestEntity1)(nil), (*fwTestEntity2)(nil), (*fwTestEntity3)(nil), (*fwTestEntity4)(nil))

	repository, _ := d3Orm.MakeSession().MakeRepository((*fwTestEntity1)(nil))

	q := repository.CreateQuery()
	_ = q.AndWhere("test_entity_1.id = ?", 1).With("d3/test/integration/relation/fwTestEntity2")
	_ = q.With("d3/test/integration/relation/fwTestEntity3")
	_ = q.With("d3/test/integration/relation/fwTestEntity4")
	entity, _ := repository.FindOne(q)

	o.Assert().Equal([]interface{}{int32(1), "entity_1_data_1"}, []interface{}{entity.(*fwTestEntity1).Id, entity.(*fwTestEntity1).Data})

	entity2 := entity.(*fwTestEntity1).Rel.Unwrap().(*fwTestEntity2)
	o.Assert().Equal([]interface{}{int32(1), "entity_2_data_1"}, []interface{}{entity2.Id, entity2.Data})
	o.Assert().Equal(3, entity2.Rel.Count())
	o.Assert().Equal(3, entity2.Rel.Get(0).(*fwTestEntity3).Rel.Count()+entity2.Rel.Get(1).(*fwTestEntity3).Rel.Count()+entity2.Rel.Get(2).(*fwTestEntity3).Rel.Count())

	o.Assert().Equal(1, wrappedDbAdapter.QueryCounter())
}
