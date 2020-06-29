package relation

import (
	"context"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type FetchWithRelationTS struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	orm       *orm.Orm
}

const d3pkg = "github.com/godzie44/d3"

func (o *FetchWithRelationTS) SetupSuite() {
	o.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

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

	o.dbAdapter = helpers.NewDbAdapterWithQueryCounter(d3pgx.NewPgxDriver(o.pgDb))
	o.orm = orm.New(o.dbAdapter)
	o.Assert().NoError(o.orm.Register(
		(*fwTestEntity1)(nil),
		(*fwTestEntity2)(nil),
		(*fwTestEntity3)(nil),
		(*fwTestEntity4)(nil),
	))
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

func (o *FetchWithRelationTS) TearDownTest() {
	o.dbAdapter.ResetCounters()
}

func TestFetchWithRelationTestSuite(t *testing.T) {
	suite.Run(t, new(FetchWithRelationTS))
}

func (o *FetchWithRelationTS) TestFetchWithOneToOne() {
	ctx := o.orm.CtxWithSession(context.Background())

	repository, _ := o.orm.MakeRepository((*fwTestEntity1)(nil))
	q := repository.Select()
	_ = q.AndWhere("test_entity_1.id = ?", 1).With(d3pkg + "/tests/integration/relation/fwTestEntity2")
	entity, _ := repository.FindOne(ctx, q)

	o.Assert().IsType(&fwTestEntity1{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity1).Id)
	o.Assert().Equal("entity_1_data_1", entity.(*fwTestEntity1).Data)

	relatedEntity := entity.(*fwTestEntity1).Rel.Unwrap().(*fwTestEntity2)
	o.Assert().Subset(
		[]interface{}{int32(1), "entity_2_data_1"},
		[]interface{}{relatedEntity.Id, relatedEntity.Data},
	)

	o.Assert().Equal(1, o.dbAdapter.QueryCounter())
}

func (o *FetchWithRelationTS) TestFetchWithOneToMany() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*fwTestEntity2)(nil))
	q := repository.Select()
	_ = q.AndWhere("test_entity_2.id = ?", 1).With(d3pkg + "/tests/integration/relation/fwTestEntity3")
	entity, _ := repository.FindOne(ctx, q)

	o.Assert().IsType(&fwTestEntity2{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity2).Id)
	o.Assert().Equal("entity_2_data_1", entity.(*fwTestEntity2).Data)

	relatedEntities := entity.(*fwTestEntity2).Rel
	o.Assert().Equal(3, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"entity_3_data_1", "entity_3_data_2", "entity_3_data_3"},
		[]string{relatedEntities.Get(0).(*fwTestEntity3).Data, relatedEntities.Get(1).(*fwTestEntity3).Data, relatedEntities.Get(2).(*fwTestEntity3).Data},
	)

	o.Assert().Equal(1, o.dbAdapter.QueryCounter())
}

func (o *FetchWithRelationTS) TestFetchWithManyToMany() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*fwTestEntity3)(nil))

	q := repository.Select()
	_ = q.AndWhere("test_entity_3.id = ?", 1).With(d3pkg + "/tests/integration/relation/fwTestEntity4")
	entity, _ := repository.FindOne(ctx, q)

	o.Assert().IsType(&fwTestEntity3{}, entity)
	o.Assert().Equal(int32(1), entity.(*fwTestEntity3).Id)
	o.Assert().Equal("entity_3_data_1", entity.(*fwTestEntity3).Data)

	relatedEntities := entity.(*fwTestEntity3).Rel
	o.Assert().Equal(2, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"entity_4_data_1", "entity_4_data_2"},
		[]string{relatedEntities.Get(0).(*fwTestEntity4).Data, relatedEntities.Get(1).(*fwTestEntity4).Data},
	)

	o.Assert().Equal(1, o.dbAdapter.QueryCounter())
}

func (o *FetchWithRelationTS) TestFetchFullGraph() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*fwTestEntity1)(nil))

	q := repository.Select()
	_ = q.AndWhere("test_entity_1.id = ?", 1).With(d3pkg + "/tests/integration/relation/fwTestEntity2")
	_ = q.With(d3pkg + "/tests/integration/relation/fwTestEntity3")
	_ = q.With(d3pkg + "/tests/integration/relation/fwTestEntity4")
	entity, _ := repository.FindOne(ctx, q)

	o.Assert().Equal([]interface{}{int32(1), "entity_1_data_1"}, []interface{}{entity.(*fwTestEntity1).Id, entity.(*fwTestEntity1).Data})

	entity2 := entity.(*fwTestEntity1).Rel.Unwrap().(*fwTestEntity2)
	o.Assert().Equal([]interface{}{int32(1), "entity_2_data_1"}, []interface{}{entity2.Id, entity2.Data})
	o.Assert().Equal(3, entity2.Rel.Count())
	o.Assert().Equal(3, entity2.Rel.Get(0).(*fwTestEntity3).Rel.Count()+entity2.Rel.Get(1).(*fwTestEntity3).Rel.Count()+entity2.Rel.Get(2).(*fwTestEntity3).Rel.Count())

	o.Assert().Equal(1, o.dbAdapter.QueryCounter())
}
