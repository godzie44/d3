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
	pgConn *pgx.Conn
	orm    *orm.Orm
	driver *helpers.DbAdapterWithQueryCounter
}

func (qts *QueryTS) SetupSuite() {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	qts.NoError(err)

	qts.pgConn = driver.UnwrapConn().(*pgx.Conn)

	qts.driver = helpers.NewDbAdapterWithQueryCounter(driver)
	qts.orm = orm.New(qts.driver)
	qts.Assert().NoError(qts.orm.Register(
		(*User)(nil),
		(*Photo)(nil),
	))

	sql, err := qts.orm.GenerateSchema()
	qts.Assert().NoError(err)

	_, err = qts.pgConn.Exec(context.Background(), sql)
	qts.Assert().NoError(err)

	_, err = qts.pgConn.Exec(context.Background(), `
INSERT INTO q_user(name, age) VALUES ('Joe', 21);
INSERT INTO q_user(name, age) VALUES ('Sara', 19);
INSERT INTO q_user(name, age) VALUES ('Piter', 33);
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
	qts.Assert().NoError(err)
}

func (qts *QueryTS) TearDownSuite() {
	_, err := qts.pgConn.Exec(context.Background(), `
DROP TABLE q_user;
DROP TABLE q_photo;
`)
	qts.Assert().NoError(err)
}

func (qts *QueryTS) TearDownTest() {
	qts.driver.ResetCounters()
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTS))
}

func (qts *QueryTS) TestQueryAll() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	allUsers, err := rep.FindAll(ctx, rep.Select())
	qts.Assert().NoError(err)

	qts.Assert().Equal(9, allUsers.Count())
}

func (qts *QueryTS) TestQueryAndWhere() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.Select().Where("age", "=", 19).AndWhere("name", "=", "Sara"))
	qts.Assert().NoError(err)

	qts.Assert().Equal(1, users.Count())
}

func (qts *QueryTS) TestQueryOrWhere() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.Select().Where("age", "=", 19).OrWhere("name", "=", "Sara"))
	qts.Assert().NoError(err)

	qts.Assert().Equal(3, users.Count())
}

func (qts *QueryTS) TestQueryNestedWhere() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.Select().Where("age", "=", 19).OrNestedWhere(func(q *query.Query) {
		q.Where("name", "=", "Sara").AndWhere("age", ">", 35)
	}))
	qts.Assert().NoError(err)

	qts.Assert().Equal(2, users.Count())
}

func (qts *QueryTS) TestQueryUnion() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	users, err := rep.FindAll(ctx, rep.Select().Where("age", "=", 19).Union(rep.Select().Where("name", "=", "Sara")))
	qts.Assert().NoError(err)

	qts.Assert().Equal(3, users.Count())
}

func (qts *QueryTS) TestQueryLimit() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	allUsers, err := rep.FindAll(ctx, rep.Select().Limit(5))
	qts.Assert().NoError(err)

	qts.Assert().Equal(5, allUsers.Count())
}

func (qts *QueryTS) TestQueryOffset() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	user, err := rep.FindOne(ctx, rep.Select().OrderBy("age ASC").Offset(1).Limit(1))
	qts.Assert().NoError(err)

	qts.Assert().Equal(21, user.(*User).age)
}

func (qts *QueryTS) TestQueryOrderBy() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	allUsersASC, err := rep.FindAll(ctx, rep.Select().OrderBy("age ASC"))
	qts.Assert().NoError(err)

	prevAge := 0
	iter := allUsersASC.MakeIter()
	for iter.Next() {
		user := iter.Value().(*User)
		qts.Assert().GreaterOrEqual(user.age, prevAge)
		prevAge = user.age
	}

	allUsersDESC, err := rep.FindAll(ctx, rep.Select().OrderBy("age DESC", "name ASC"))
	qts.Assert().NoError(err)

	prevAge = math.MaxInt64
	iter = allUsersDESC.MakeIter()
	for iter.Next() {
		user := iter.Value().(*User)
		qts.Assert().LessOrEqual(user.age, prevAge)
		prevAge = user.age
	}
}

func (qts *QueryTS) TestQueryWith() {
	ctx := qts.orm.CtxWithSession(context.Background())
	rep, err := qts.orm.MakeRepository((*User)(nil))
	qts.Assert().NoError(err)

	q := rep.Select()
	err = q.With("github.com/godzie44/d3/tests/integration/query/Photo")
	qts.Assert().NoError(err)

	users, err := rep.FindAll(ctx, q.Where("q_photo.user_id", "IS NOT NULL").OrderBy("age ASC"))
	qts.Assert().NoError(err)

	qts.Assert().Equal(3, users.Count())
	qts.Assert().Equal(2, users.Get(0).(*User).photos.Count())

	qts.Assert().Equal(1, qts.driver.QueryCounter())
}

func (qts *QueryTS) TestQueryJoin() {
	session := qts.orm.MakeSession()

	q := query.New().Select("*").From("q_user").Join(query.JoinInner, "q_photo", "q_user.id=q_photo.user_id")

	result, err := session.Execute(q)
	qts.Assert().NoError(err)

	qts.Assert().Len(result, 5)
}

func (qts *QueryTS) TestQueryGroupBy() {
	session := qts.orm.MakeSession()

	q := query.New().Select("age", "count(*)").From("q_user").GroupBy("age")

	result, err := session.Execute(q)
	qts.Assert().NoError(err)

	qts.Assert().Len(result, 7)
	for _, res := range result {
		if res["age"] == 42 || res["age"] == 33 {
			qts.Assert().Equal(2, res["count(*)"])
		}
	}
}

func (qts *QueryTS) TestQueryHaving() {
	session := qts.orm.MakeSession()

	q := query.New().Select("age").From("q_user").GroupBy("age").Having("count(*)", ">", 1)

	result, err := session.Execute(q)
	qts.Assert().NoError(err)

	qts.Assert().Len(result, 2)
}
