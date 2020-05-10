package cache

import (
	"context"
	"d3/adapter"
	"d3/orm"
	entity3 "d3/orm/entity"
	"d3/test/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type IMCacheTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *IMCacheTS) SetupSuite() {
	o.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS im_test_entity_1(
		id integer NOT NULL,
		data text NOT NULL,
		CONSTRAINT im_test_entity_t1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS im_test_entity_2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t1_id integer,
		CONSTRAINT im_test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS im_test_entity_3(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t2_id integer,
		CONSTRAINT im_test_entity_t3_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO im_test_entity_1(id, data) VALUES (1, 'entity_1_data');
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (1, 'entity_2_data_1', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (2, 'entity_2_data_2', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (3, 'entity_2_data_3', 1);
INSERT INTO im_test_entity_3(id, data, t2_id) VALUES (1, 'entity_3_data', 1);
`)
	o.Assert().NoError(err)

}

func (o *IMCacheTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE im_test_entity_1;
DROP TABLE im_test_entity_2;
DROP TABLE im_test_entity_3;
`)
	o.Assert().NoError(err)
}

type entity1 struct {
	entity struct{}           `d3:"table_name:im_test_entity_1"` //nolint:unused,structcheck
	Id     int32              `d3:"pk:auto"`
	Rel    entity3.Collection `d3:"one_to_many:<target_entity:d3/test/integration/cache/entity2,join_on:t1_id>,type:eager"`
	Data   string
}

type entity2 struct {
	entity struct{} `d3:"table_name:im_test_entity_2"` //nolint:unused,structcheck
	Id     int32    `d3:"pk:auto"`
	Data   string
}

func TestIdentityMapCacheSuite(t *testing.T) {
	suite.Run(t, new(IMCacheTS))
}

func (o *IMCacheTS) TestNoQueryCreateForCachedEntities() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(wrappedDbAdapter)
	_ = d3Orm.Register((*entity1)(nil), (*entity2)(nil))

	session := d3Orm.MakeSession()
	repository, _ := session.MakeRepository((*entity1)(nil))
	_, err := repository.FindOne(repository.CreateQuery().AndWhere("im_test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())

	_, err = repository.FindOne(repository.CreateQuery().AndWhere("im_test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())

	repository2, _ := session.MakeRepository((*entity2)(nil))
	_, err = repository2.FindOne(repository2.CreateQuery().AndWhere("im_test_entity_2.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())
}
