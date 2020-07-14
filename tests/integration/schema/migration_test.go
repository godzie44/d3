package schema

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

type MigrationTestSuite struct {
	suite.Suite
	pgConn *pgx.Conn
	orm    *orm.Orm
}

func (m *MigrationTestSuite) SetupSuite() {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	m.NoError(err)

	m.pgConn = driver.UnwrapConn().(*pgx.Conn)

	m.orm = orm.New(driver)
	m.Assert().NoError(m.orm.Register(
		(*shop)(nil),
		(*profile)(nil),
		(*book)(nil),
		(*author)(nil),
	))
}

func (m *MigrationTestSuite) TearDownSuite() {
	_, err := m.pgConn.Exec(context.Background(), `
DROP TABLE IF EXISTS shop_m;
DROP TABLE IF EXISTS profile_m;
DROP TABLE IF EXISTS author_m;
DROP TABLE IF EXISTS book_m;
DROP TABLE IF EXISTS book_author_m;
`)
	m.Assert().NoError(err)
}

func (m *MigrationTestSuite) TestCreateSchema() {
	sql, err := m.orm.GenerateSchema()
	m.NoError(err)

	_, err = m.pgConn.Exec(context.Background(), sql)
	m.NoError(err)

	helpers.NewPgTester(m.T(), m.pgConn).
		SeeTable("shop_m").
		SeeTable("profile_m").
		SeeTable("author_m").
		SeeTable("book_m").
		SeeTable("book_author_m")
}

func TestMigrationTs(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
