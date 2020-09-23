package schema

import (
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MigrationTestSuite struct {
	suite.Suite
	tester    helpers.DBTester
	orm       *orm.Orm
	execSqlFn func(sql string) error
}

func (m *MigrationTestSuite) SetupSuite() {
	m.Assert().NoError(m.orm.Register(
		(*shop)(nil),
		(*profile)(nil),
		(*book)(nil),
		(*author)(nil),
	))
}

func (m *MigrationTestSuite) TearDownSuite() {
	m.NoError(m.execSqlFn(`
DROP TABLE IF EXISTS shop_m;
DROP TABLE IF EXISTS profile_m;
DROP TABLE IF EXISTS author_m;
DROP TABLE IF EXISTS book_m;
DROP TABLE IF EXISTS book_author_m;

DROP INDEX IF EXISTS shop_name_idx;
DROP INDEX IF EXISTS book_name_idx;
DROP INDEX IF EXISTS author_name_idx;
DROP INDEX IF EXISTS book_isbn_idx;
`))
}

func (m *MigrationTestSuite) TestCreateSchema() {
	sql, err := m.orm.GenerateSchema()
	m.NoError(err)

	m.NoError(m.execSqlFn(sql))

	m.tester.
		SeeTable("shop_m").
		SeeTable("profile_m").
		SeeTable("author_m").
		SeeTable("book_m").
		SeeTable("book_author_m").
		SeeIndex("shop_m", "shop_name_idx", false).
		SeeIndex("book_m", "book_name_idx", false).
		SeeIndex("book_m", "book_isbn_idx", true).
		SeeIndex("author_m", "author_name_idx", false)
}

func TestPGMigrationTs(t *testing.T) {
	_, d3orm, execSqlFn, tester := db.CreatePGTestComponents(t)

	suite.Run(t, &MigrationTestSuite{
		orm:       d3orm,
		tester:    tester,
		execSqlFn: execSqlFn,
	})
}

func TestSqliteMigrationTs(t *testing.T) {
	_, d3orm, execSqlFn, tester := db.CreateSQLiteTestComponents(t, "_migrations")

	suite.Run(t, &MigrationTestSuite{
		orm:       d3orm,
		tester:    tester,
		execSqlFn: execSqlFn,
	})
}
