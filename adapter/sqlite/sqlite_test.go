package sqlite

import (
	"database/sql"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type sqliteDriverTS struct {
	suite.Suite
	driver *sqliteDriver
	tester *helpers.SqliteTester
}

const dbName = "./test.db"

func (s *sqliteDriverTS) SetupTest() {
	_, err := s.driver.UnwrapConn().(*sql.DB).Exec(`CREATE TABLE IF NOT EXISTS d3_test_table(
		id integer NOT NULL,
		data text NOT NULL
	)`)
	s.NoError(err)

	_, err = s.driver.UnwrapConn().(*sql.DB).Exec(`
INSERT INTO d3_test_table(id, data) VALUES (1, 'test 1');
INSERT INTO d3_test_table(id, data) VALUES (2, 'test 2');
INSERT INTO d3_test_table(id, data) VALUES (3, 'test 3');
`)
	s.NoError(err)
}

func (s *sqliteDriverTS) TearDownTest() {
	_, err := s.driver.UnwrapConn().(*sql.DB).Exec(`DROP TABLE d3_test_table;`)
	s.NoError(err)
}

func (s *sqliteDriverTS) TearDownSuite() {
	s.NoError(s.driver.Close())
	s.NoError(os.Remove(dbName))
}

func TestPgxConnDriverTestSuite(t *testing.T) {
	driver, err := NewSqliteDriver(dbName)
	assert.NoError(t, err)

	suite.Run(t, &sqliteDriverTS{
		driver: driver,
		tester: helpers.NewSqliteTester(t, driver.UnwrapConn().(*sql.DB)),
	})
}

func (s *sqliteDriverTS) TestPgxDriverQuery() {
	tx, err := s.driver.BeginTx()
	s.NoError(err)
	defer tx.Commit()

	data, err := s.driver.ExecuteQuery(query.New().Select("*").From("d3_test_table").Where("id", "=", "1"), tx)
	s.NoError(err)

	s.Len(data, 1)
	s.Equal(data[0], map[string]interface{}{"id": int64(1), "data": "test 1"})

	data, err = s.driver.ExecuteQuery(query.New().Select("*").From("d3_test_table"), tx)
	s.NoError(err)

	s.Len(data, 3)
}

func (s *sqliteDriverTS) TestSqliteDriverTxInsert() {
	tx, err := s.driver.BeginTx()
	s.NoError(err)

	pusher := s.driver.MakePusher(tx)

	err = pusher.Insert("d3_test_table", []string{"id", "data"}, []interface{}{4, "test 4"}, persistence.Undefined)
	s.NoError(err)
	s.NoError(tx.Commit())

	s.tester.SeeFour("select * from d3_test_table")
}

func (s *sqliteDriverTS) TestSqliteDriverTxUpdate() {
	tx, err := s.driver.BeginTx()
	s.NoError(err)

	pusher := s.driver.MakePusher(tx)

	err = pusher.Update("d3_test_table", []string{"data"}, []interface{}{"test upd"}, map[string]interface{}{"id": 1})
	s.NoError(err)
	s.NoError(tx.Commit())

	s.tester.SeeOne("select * from d3_test_table where data = 'test upd'")
}

func (s *sqliteDriverTS) TestSqliteDriverTxDelete() {
	tx, err := s.driver.BeginTx()
	s.NoError(err)

	pusher := s.driver.MakePusher(tx)

	err = pusher.Remove("d3_test_table", map[string]interface{}{"id": 1})
	s.NoError(err)
	s.NoError(tx.Commit())

	s.tester.SeeTwo("select * from d3_test_table")
}

func (s *sqliteDriverTS) TestSqliteDriverTxRollback() {
	tx, err := s.driver.BeginTx()
	s.NoError(err)

	pusher := s.driver.MakePusher(tx)

	err = pusher.Insert("d3_test_table", []string{"id", "data"}, []interface{}{4, "test 4"}, persistence.Undefined)
	s.NoError(err)
	s.NoError(tx.Rollback())

	s.tester.SeeThree("select * from d3_test_table")
}
