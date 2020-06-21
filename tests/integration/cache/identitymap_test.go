package cache

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

	_, err = o.pgDb.Exec(context.Background(), `
INSERT INTO im_test_entity_1(id, data) VALUES (1, 'entity_1_data');
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (1, 'entity_2_data_1', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (2, 'entity_2_data_2', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (3, 'entity_2_data_3', 1);
`)
	o.Assert().NoError(err)

}

func (o *IMCacheTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE im_test_entity_1;
DROP TABLE im_test_entity_2;
`)
	o.Assert().NoError(err)
}

func TestIdentityMapCacheSuite(t *testing.T) {
	suite.Run(t, new(IMCacheTS))
}

func (o *IMCacheTS) TestNoQueryCreateForCachedEntities() {
	wrappedDbAdapter := helpers.NewDbAdapterWithQueryCounter(d3pgx.NewPgxDriver(o.pgDb))
	d3Orm := orm.NewOrm(wrappedDbAdapter)
	err := d3Orm.Register(
		(*entity1)(nil),
		(*entity2)(nil),
	)
	o.NoError(err)

	ctx := d3Orm.CtxWithSession(context.Background())
	repository, _ := d3Orm.MakeRepository((*entity1)(nil))
	_, err = repository.FindOne(ctx, repository.MakeQuery().AndWhere("im_test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())

	_, err = repository.FindOne(ctx, repository.MakeQuery().AndWhere("im_test_entity_1.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())

	repository2, _ := d3Orm.MakeRepository((*entity2)(nil))
	_, err = repository2.FindOne(ctx, repository2.MakeQuery().AndWhere("im_test_entity_2.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, wrappedDbAdapter.QueryCounter())
}
