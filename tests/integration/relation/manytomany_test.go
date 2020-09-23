package relation

import (
	"context"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ManyToManyRelationTS struct {
	suite.Suite
	orm       *orm.Orm
	execSqlFn func(sql string) error
}

func (m *ManyToManyRelationTS) SetupSuite() {
	err := m.execSqlFn(`CREATE TABLE IF NOT EXISTS book(
		id integer NOT NULL,
		name text NOT NULL,
		CONSTRAINT book_pkey PRIMARY KEY (id)
	)`)
	m.NoError(err)

	err = m.execSqlFn(`CREATE TABLE IF NOT EXISTS author(
		id integer NOT NULL,
		name character varying(200) NOT NULL,
		CONSTRAINT author_pkey PRIMARY KEY (id)
	)`)
	m.NoError(err)

	err = m.execSqlFn(`CREATE TABLE IF NOT EXISTS book_author(
		book_id integer NOT NULL,
		author_id integer NOT NULL
	)`)
	m.NoError(err)

	err = m.execSqlFn(`CREATE TABLE IF NOT EXISTS redactor(
		id integer NOT NULL,
		name character varying(200) NOT NULL,
		CONSTRAINT redactor_pkey PRIMARY KEY (id)
	)`)
	m.NoError(err)

	err = m.execSqlFn(`CREATE TABLE IF NOT EXISTS author_redactor(
		author_id integer NOT NULL,
		redactor_id integer NOT NULL
	)`)
	m.NoError(err)

	err = m.execSqlFn(`
INSERT INTO book(id, name) VALUES (1, 'Antic Hay');
INSERT INTO book(id, name) VALUES (2, 'An Evil Cradling');
INSERT INTO book(id, name) VALUES (3, 'Arms and the Man');
INSERT INTO author(id, name) VALUES (1, 'Aldous Huxley');
INSERT INTO author(id, name) VALUES (2, 'Brian Keenan');
INSERT INTO redactor(id, name) VALUES (1, 'George Bernard Shaw');
INSERT INTO book_author(book_id, author_id) VALUES (1, 1);
INSERT INTO book_author(book_id, author_id) VALUES (1, 2);
INSERT INTO book_author(book_id, author_id) VALUES (2, 2);
INSERT INTO book_author(book_id, author_id) VALUES (3, 1);
INSERT INTO author_redactor(author_id, redactor_id) VALUES (1, 1);
`)
	m.NoError(err)

	m.NoError(m.orm.Register(
		(*BookLL)(nil),
		(*AuthorLL)(nil),
		(*Redactor)(nil),
		(*BookEL)(nil),
		(*AuthorEL)(nil),
	))
}

func (m *ManyToManyRelationTS) TearDownSuite() {
	m.NoError(m.execSqlFn(`
DROP TABLE book;
DROP TABLE author;
DROP TABLE redactor;
DROP TABLE book_author;
DROP TABLE author_redactor;
`))
}

func TestPGManyToManyTestSuite(t *testing.T) {
	_, d3orm, execSqlFn, _ := db.CreatePGTestComponents(t)

	mtmTS := &ManyToManyRelationTS{
		orm:       d3orm,
		execSqlFn: execSqlFn,
	}
	suite.Run(t, mtmTS)
}

func TestSQLiteManyToManyTestSuite(t *testing.T) {
	_, d3orm, execSqlFn, _ := db.CreateSQLiteTestComponents(t, "_m_to_m")

	mtmTS := &ManyToManyRelationTS{
		orm:       d3orm,
		execSqlFn: execSqlFn,
	}
	suite.Run(t, mtmTS)
}

func (m *ManyToManyRelationTS) TestLazyRelation() {
	ctx := m.orm.CtxWithSession(context.Background())
	repository, err := m.orm.MakeRepository((*BookLL)(nil))
	m.Assert().NoError(err)

	entity, err := repository.FindOne(ctx, repository.Select().Where("book.id", "=", 1))
	m.Assert().NoError(err)

	m.Assert().IsType(&BookLL{}, entity)
	m.Assert().Equal(int32(1), entity.(*BookLL).ID)
	m.Assert().Equal("Antic Hay", entity.(*BookLL).Name)

	relatedEntities := entity.(*BookLL).Authors
	m.Assert().Equal(relatedEntities.Count(), 2)
	m.Assert().Subset(
		[]string{"Aldous Huxley", "Brian Keenan"},
		[]string{relatedEntities.Get(0).(*AuthorLL).Name, relatedEntities.Get(1).(*AuthorLL).Name},
	)
}

func (m *ManyToManyRelationTS) TestEagerRelation() {
	ctx := m.orm.CtxWithSession(context.Background())
	repository, err := m.orm.MakeRepository((*BookEL)(nil))
	m.Assert().NoError(err)

	entity, err := repository.FindOne(ctx, repository.Select().Where("book.id", "=", 1))
	m.Assert().NoError(err)

	m.Assert().IsType(&BookEL{}, entity)
	m.Assert().Equal(int32(1), entity.(*BookEL).Id)
	m.Assert().Equal("Antic Hay", entity.(*BookEL).Name)

	relatedEntities := entity.(*BookEL).Rel
	m.Assert().Equal(2, relatedEntities.Count())
	m.Assert().Subset(
		[]string{"Aldous Huxley", "Brian Keenan"},
		[]string{relatedEntities.Get(0).(*AuthorEL).Name, relatedEntities.Get(1).(*AuthorEL).Name},
	)

	if relatedEntities.Get(0).(*AuthorEL).Rel.Count() != 1 && relatedEntities.Get(1).(*AuthorEL).Rel.Count() != 1 {
		m.Assert().Fail("testEntity3 not found")
	}
}
