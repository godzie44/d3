package schema

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type MigrationTestSuite struct {
	suite.Suite
	pgDb *pgx.Conn
	orm  *orm.Orm
}

func (m *MigrationTestSuite) SetupSuite() {
	m.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	m.orm = orm.NewOrm(adapter.NewGoPgXAdapter(m.pgDb, &adapter.SquirrelAdapter{}))
	m.Assert().NoError(m.orm.Register(
		(*shop)(nil),
		(*profile)(nil),
		(*book)(nil),
		(*author)(nil),
	))
}

func (m *MigrationTestSuite) TearDownSuite() {
	_, err := m.pgDb.Exec(context.Background(), `
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

	_, err = m.pgDb.Exec(context.Background(), sql)
	m.NoError(err)

	helpers.NewPgTester(m.T(), m.pgDb).
		SeeTable("shop_m").
		SeeTable("profile_m").
		SeeTable("author_m").
		SeeTable("book_m").
		SeeTable("book_author_m")
}

func TestMigrationTs(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
