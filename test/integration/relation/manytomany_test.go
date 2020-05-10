package relation

import (
	"context"
	"d3/adapter"
	"d3/orm"
	entity2 "d3/orm/entity"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ManyToManyRelationTS struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (o *ManyToManyRelationTS) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	o.pgDb, _ = pgx.Connect(context.Background(), dsn)

	_, err := o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book(
		id integer NOT NULL,
		name text NOT NULL,
		CONSTRAINT book_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS author(
		id integer NOT NULL,
		name character varying(200) NOT NULL,
		CONSTRAINT author_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book_author(
		book_id integer NOT NULL,
		author_id integer NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS redactor(
		id integer NOT NULL,
		name character varying(200) NOT NULL,
		CONSTRAINT redactor_pkey PRIMARY KEY (id)
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS author_redactor(
		author_id integer NOT NULL,
		redactor_id integer NOT NULL
	)`)
	o.Assert().NoError(err)

	_, err = o.pgDb.Exec(context.Background(), `
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
	o.Assert().NoError(err)
}

func (o *ManyToManyRelationTS) TearDownSuite() {
	_, err := o.pgDb.Exec(context.Background(), `
DROP TABLE book;
DROP TABLE author;
DROP TABLE redactor;
DROP TABLE book_author;
DROP TABLE author_redactor;
`)
	o.Assert().NoError(err)
}

func TestManyToManyTestSuite(t *testing.T) {
	suite.Run(t, new(ManyToManyRelationTS))
}

type BookLL struct {
	entity  struct{}           `d3:"table_name:book"` //nolint:unused,structcheck
	ID      int32              `d3:"pk:auto"`
	Authors entity2.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/AuthorLL,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	Name    string
}

type AuthorLL struct {
	entity struct{} `d3:"table_name:author"` //nolint:unused,structcheck
	ID     int32    `d3:"pk:auto"`
	Name   string
}

func (o *ManyToManyRelationTS) TestLazyRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*BookLL)(nil), (*AuthorLL)(nil), (*Redactor)(nil))
	o.Assert().NoError(err)

	session := d3Orm.MakeSession()
	repository, err := session.MakeRepository((*BookLL)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("book.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&BookLL{}, entity)
	o.Assert().Equal(int32(1), entity.(*BookLL).ID)
	o.Assert().Equal("Antic Hay", entity.(*BookLL).Name)

	relatedEntities := entity.(*BookLL).Authors
	o.Assert().Equal(relatedEntities.Count(), 2)
	o.Assert().Subset(
		[]string{"Aldous Huxley", "Brian Keenan"},
		[]string{relatedEntities.Get(0).(*AuthorLL).Name, relatedEntities.Get(1).(*AuthorLL).Name},
	)
}

type BookEL struct {
	entity struct{}           `d3:"table_name:book"` //nolint:unused,structcheck
	Id     int32              `d3:"pk:auto"`
	Rel    entity2.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/AuthorEL,join_on:book_id,reference_on:author_id,join_table:book_author>,type:eager"`
	Name   string
}

type AuthorEL struct {
	entity struct{}           `d3:"table_name:author"` //nolint:unused,structcheck
	Id     int32              `d3:"pk:auto"`
	Rel    entity2.Collection `d3:"many_to_many:<target_entity:d3/test/integration/relation/Redactor,join_on:author_id,reference_on:redactor_id,join_table:author_redactor>,type:eager"`
	Name   string
}

type Redactor struct {
	entity struct{} `d3:"table_name:redactor"` //nolint:unused,structcheck
	Id     int32    `d3:"pk:auto"`
	Name   string
}

func (o *ManyToManyRelationTS) TestEagerRelation() {
	d3Orm := orm.NewOrm(adapter.NewGoPgXAdapter(o.pgDb, &adapter.SquirrelAdapter{}))
	err := d3Orm.Register((*BookEL)(nil), (*AuthorEL)(nil), (*Redactor)(nil))
	o.Assert().NoError(err)

	session := d3Orm.MakeSession()
	repository, err := session.MakeRepository((*BookEL)(nil))
	o.Assert().NoError(err)

	entity, err := repository.FindOne(repository.CreateQuery().AndWhere("book.id = ?", 1))
	o.Assert().NoError(err)

	o.Assert().IsType(&BookEL{}, entity)
	o.Assert().Equal(int32(1), entity.(*BookEL).Id)
	o.Assert().Equal("Antic Hay", entity.(*BookEL).Name)

	relatedEntities := entity.(*BookEL).Rel
	o.Assert().Equal(2, relatedEntities.Count())
	o.Assert().Subset(
		[]string{"Aldous Huxley", "Brian Keenan"},
		[]string{relatedEntities.Get(0).(*AuthorEL).Name, relatedEntities.Get(1).(*AuthorEL).Name},
	)

	if relatedEntities.Get(0).(*AuthorEL).Rel.Count() != 1 && relatedEntities.Get(1).(*AuthorEL).Rel.Count() != 1 {
		o.Assert().Fail("testEntity3 not found")
	}
}
