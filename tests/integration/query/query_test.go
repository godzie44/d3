package query

import (
	"context"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/query"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"math"
	"os"
	"testing"
)

type QueryTS struct {
	suite.Suite
	pgDb   *pgx.Conn
	orm    *orm.Orm
	driver *helpers.DbAdapterWithQueryCounter
}

func (q *QueryTS) SetupSuite() {
	q.pgDb, _ = pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	q.driver = helpers.NewDbAdapterWithQueryCounter(d3pgx.NewPgxDriver(q.pgDb))
	q.orm = orm.New(q.driver)
	q.Assert().NoError(q.orm.Register(
		(*User)(nil),
		(*Photo)(nil),
	))

	sql, err := q.orm.GenerateSchema()
	q.Assert().NoError(err)

	_, err = q.pgDb.Exec(context.Background(), sql)
	q.Assert().NoError(err)

	_, err = q.pgDb.Exec(context.Background(), `
INSERT INTO q_user(name, age) VALUES ('Joe', 21);
INSERT INTO q_user(name, age) VALUES ('Sara', 19);
INSERT INTO q_user(name, age) VALUES ('Piter', 31);
INSERT INTO q_user(name, age) VALUES ('Victor', 41);
INSERT INTO q_user(name, age) VALUES ('Emili', 41);
INSERT INTO q_user(name, age) VALUES ('Sara', 42);
INSERT INTO q_user(name, age) VALUES ('Sara', 33);
INSERT INTO q_user(name, age) VALUES ('John', 29);
INSERT INTO q_user(name, age) VALUES ('John', 26);
INSERT INTO q_photo(user_id, src) VALUES (2, 'http://sara_pic_url');
INSERT INTO q_photo(user_id, src) VALUES (2, 'http://sara_pic_url');
INSERT INTO q_photo(user_id, src) VALUES (4, 'http://victor_pic_url');
INSERT INTO q_photo(user_id, src) VALUES (5, 'http://emili_pic_url');
INSERT INTO q_photo(user_id, src) VALUES (5, 'http://emili_pic_url');
`)
	q.Assert().NoError(err)
}

func (q *QueryTS) TearDownSuite() {
	_, err := q.pgDb.Exec(context.Background(), `
DROP TABLE q_user;
DROP TABLE q_photo;
`)
	q.Assert().NoError(err)
}

func (q *QueryTS) TearDownTest() {
	q.driver.ResetCounters()
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTS))
}

func (q *QueryTS) TestQueryAll() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	allUsers, err := rep.FindAll(ctx, rep.MakeQuery())
	q.Assert().NoError(err)

	q.Assert().Equal(9, allUsers.Count())
}

func (q *QueryTS) TestQueryAndWhere() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.MakeQuery().AndWhere("age = ?", 19))
	q.Assert().NoError(err)

	q.Assert().Equal(1, users.Count())
}

func (q *QueryTS) TestQueryOrWhere() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.MakeQuery().AndWhere("age = ?", 19).OrWhere("name = ?", "Sara"))
	q.Assert().NoError(err)

	q.Assert().Equal(3, users.Count())
}

func (q *QueryTS) TestQueryUnion() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.MakeQuery().AndWhere("age = ?", 19).Union(rep.MakeQuery().AndWhere("name = ?", "Sara")))
	q.Assert().NoError(err)

	q.Assert().Equal(3, users.Count())
}

func (q *QueryTS) TestQueryLimit() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	allUsers, err := rep.FindAll(ctx, rep.MakeQuery().Limit(5))
	q.Assert().NoError(err)

	q.Assert().Equal(5, allUsers.Count())
}

func (q *QueryTS) TestQueryOffset() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	user, err := rep.FindOne(ctx, rep.MakeQuery().OrderBy("age ASC").Offset(1).Limit(1))
	q.Assert().NoError(err)

	q.Assert().Equal(21, user.(*User).age)
}

func (q *QueryTS) TestQueryOrderBy() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	allUsersASC, err := rep.FindAll(ctx, rep.MakeQuery().OrderBy("age ASC"))
	q.Assert().NoError(err)

	prevAge := 0
	iter := allUsersASC.MakeIter()
	for iter.Next() {
		user := iter.Value().(*User)
		q.Assert().GreaterOrEqual(user.age, prevAge)
		prevAge = user.age
	}

	allUsersDESC, err := rep.FindAll(ctx, rep.MakeQuery().OrderBy("age DESC", "name ASC"))
	q.Assert().NoError(err)

	prevAge = math.MaxInt64
	iter = allUsersDESC.MakeIter()
	for iter.Next() {
		user := iter.Value().(*User)
		q.Assert().LessOrEqual(user.age, prevAge)
		prevAge = user.age
	}
}

func (q *QueryTS) TestQueryWith() {
	ctx := q.orm.CtxWithSession(context.Background())
	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	query := rep.MakeQuery()
	err = query.With("github.com/godzie44/d3/tests/integration/query/Photo")
	q.Assert().NoError(err)

	users, err := rep.FindAll(ctx, query.AndWhere("q_photo.user_id IS NOT NULL").OrderBy("age ASC"))
	q.Assert().NoError(err)

	q.Assert().Equal(3, users.Count())
	q.Assert().Equal(2, users.Get(0).(*User).photos.Count())

	q.Assert().Equal(1, q.driver.QueryCounter())
}

func (q *QueryTS) TestQueryJoin() {
	ctx := q.orm.CtxWithSession(context.Background())

	rep, err := q.orm.MakeRepository((*User)(nil))
	q.Assert().NoError(err)

	query := rep.MakeQuery().Join(query.JoinInner, "q_photo", "q_user.id=q_photo.user_id")

	res, err := rep.FindAll(ctx, query)
	q.Assert().NoError(err)

	q.Assert().Equal(3, res.Count())
}
