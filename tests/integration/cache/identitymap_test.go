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
	pgConn  *pgx.Conn
	adapter *helpers.DbAdapterWithQueryCounter
	orm     *orm.Orm
}

func (o *IMCacheTS) SetupSuite() {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	o.NoError(err)

	o.pgConn = driver.UnwrapConn().(*pgx.Conn)

	_, err = o.pgConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS im_test_entity_1(
		id integer NOT NULL,
		data text NOT NULL,
		CONSTRAINT im_test_entity_t1_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS im_test_entity_2(
		id integer NOT NULL,
		data character varying(200) NOT NULL,
		t1_id integer,
		CONSTRAINT im_test_entity_t2_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgConn.Exec(context.Background(), `
INSERT INTO im_test_entity_1(id, data) VALUES (1, 'entity_1_data');
INSERT INTO im_test_entity_1(id, data) VALUES (2, 'entity_1_data');
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (1, 'entity_2_data_1', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (2, 'entity_2_data_2', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (3, 'entity_2_data_3', 1);
INSERT INTO im_test_entity_2(id, data, t1_id) VALUES (4, 'entity_2_data_4', 1);
`)
	o.Assert().NoError(err)

	o.adapter = helpers.NewDbAdapterWithQueryCounter(driver)
	o.orm = orm.New(o.adapter)
	err = o.orm.Register(
		(*entity1)(nil),
		(*entity2)(nil),
	)
	o.NoError(err)
}

func (o *IMCacheTS) TearDownSuite() {
	_, err := o.pgConn.Exec(context.Background(), `
DROP TABLE im_test_entity_1;
DROP TABLE im_test_entity_2;
`)
	o.Assert().NoError(err)
}

func (o *IMCacheTS) TearDownTest() {
	o.adapter.ResetCounters()
}

func TestIdentityMapCacheSuite(t *testing.T) {
	suite.Run(t, new(IMCacheTS))
}

func (o *IMCacheTS) TestNoDBCallForEqualQuery() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*entity1)(nil))
	_, err := repository.FindOne(ctx, repository.Select().Where("im_test_entity_1.id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, o.adapter.QueryCounter())

	_, err = repository.FindOne(ctx, repository.Select().Where("im_test_entity_1.id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, o.adapter.QueryCounter())

	repository2, _ := o.orm.MakeRepository((*entity2)(nil))
	_, err = repository2.FindOne(ctx, repository2.Select().Where("im_test_entity_2.id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, o.adapter.QueryCounter())
}

func (o *IMCacheTS) TestDBCallForEqualQueryIfKeyNotFoundInCache() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*entity1)(nil))
	_, err := repository.FindOne(ctx, repository.Select().Where("im_test_entity_1.id", "=", 1))
	o.Assert().NoError(err)

	o.Assert().Equal(2, o.adapter.QueryCounter())

	_, err = repository.FindOne(ctx, repository.Select().Where("im_test_entity_1.id", "=", 2))
	o.Assert().NoError(err)

	o.Assert().Equal(4, o.adapter.QueryCounter())
}

func (o *IMCacheTS) TestNoDBCallForInQuery() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*entity2)(nil))
	_, err := repository.FindOne(ctx, repository.Select().Where("id", "IN", 1, 2, 3))
	o.Assert().NoError(err)

	o.Assert().Equal(1, o.adapter.QueryCounter())

	_, err = repository.FindOne(ctx, repository.Select().Where("id", "IN", 2, 3))
	o.Assert().NoError(err)

	o.Assert().Equal(1, o.adapter.QueryCounter())
}

func (o *IMCacheTS) TestDBCallForInQueryIfKeyNotFoundInCache() {
	ctx := o.orm.CtxWithSession(context.Background())
	repository, _ := o.orm.MakeRepository((*entity2)(nil))
	_, err := repository.FindOne(ctx, repository.Select().Where("id", "IN", 1, 2, 3))
	o.Assert().NoError(err)

	o.Assert().Equal(1, o.adapter.QueryCounter())

	_, err = repository.FindOne(ctx, repository.Select().Where("id", "IN", 3, 4))
	o.Assert().NoError(err)

	o.Assert().Equal(2, o.adapter.QueryCounter())
}
