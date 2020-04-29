package persist

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"d3/test/helpers"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UpdateTs struct {
	suite.Suite
	pgDb *pgx.Conn
}

func (u *UpdateTs) SetupSuite() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "d3db")
	u.pgDb, _ = pgx.Connect(context.Background(), dsn)

	err := createSchema(u.pgDb)
	u.NoError(err)
}

func (u *UpdateTs) TearDownSuite() {
	u.NoError(deleteSchema(u.pgDb))
}

func (u *UpdateTs) TearDownTest() {
	u.NoError(clearSchema(u.pgDb))
}

func TestUpdateSuite(t *testing.T) {
	suite.Run(t, new(UpdateTs))
}

func (u *UpdateTs) TestInsertThenUpdate() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	u.NoError(err)

	u.NoError(session.Flush())

	shop.Name = "new shop"

	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.UpdateCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='new shop'").
		See(0, "SELECT * FROM shop_p WHERE name='shop'")
}

func (u *UpdateTs) TestInsertThenUpdateOToMRelation() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	u.NoError(err)

	u.NoError(session.Flush())

	shop.Books.Remove(0)

	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.UpdateCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NOT NULL").
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NULL")
}

func (u *UpdateTs) TestInsertThenUpdateMToMRelations() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	u.NoError(err)

	u.NoError(session.Flush())

	book := shop.Books.Get(0).(*Book)
	author := book.Authors.Get(1).(*Author)

	book.Authors.Remove(1)

	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.DeleteCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		See(0, "SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", book.Id, author.Id)
}

func (u *UpdateTs) TestInsertThenFullUpdate() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	session := d3Orm.CreateSession()

	shop, err := createAndPersistsShop(d3Orm, session)
	u.NoError(err)

	u.NoError(session.Flush())

	newProfile := &ShopProfile{Description: "new shop profile"}
	shop.Profile = entity.NewWrapEntity(newProfile)
	shop.Name = "new shop"

	shop.Books.Remove(0)

	newAuthor := &Author{Name: "new author"}
	newBook := &Book{Name: "new book", Authors: entity.NewCollection([]interface{}{newAuthor})}
	shop.Books.Add(newBook)

	oldBook := shop.Books.Get(0).(*Book)
	oldBook.Authors.Remove(1)

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.DeleteCounter())
	u.Equal(2, dbAdapter.UpdateCounter())
	u.Equal(4, dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1 and profile_id = $2", "new shop", newProfile.Id).
		SeeOne("SELECT * FROM book_p WHERE id = $1", newBook.Id).
		SeeOne("SELECT * FROM author_p WHERE id = $1", newAuthor.Id).
		SeeOne("SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", newBook.Id, newAuthor.Id)
}

func createAndPersistsShop(orm *orm.Orm, s *orm.Session) (*Shop, error) {
	repository, _ := orm.CreateRepository(s, (*Shop)(nil))

	author1 := &Author{
		Name: "author1",
	}
	shop := &Shop{
		Books: entity.NewCollection([]interface{}{&Book{
			Authors: entity.NewCollection([]interface{}{author1, &Author{Name: "author 2"}}),
			Name:    "book 1",
		}, &Book{
			Authors: entity.NewCollection([]interface{}{author1, &Author{Name: "author 3"}}),
			Name:    "book 2",
		}}),
		Profile: entity.NewWrapEntity(&ShopProfile{
			Description: "this is test shop",
		}),
		Name: "shop",
	}

	return shop, repository.Persists(shop)
}

func (u *UpdateTs) TestSelectThenSimpleUpdate() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	fillDb(u.Assert(), dbAdapter)
	session := d3Orm.CreateSession()

	repo, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1001"))
	u.NoError(err)
	shop2i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1002"))
	u.NoError(err)

	shop1i.(*Shop).Name = "new shop 1001 name"
	shop2i.(*Shop).Name = "new shop 1002 name"

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(2, dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1", "new shop 1001 name").
		SeeOne("SELECT * FROM shop_p WHERE name = $1", "new shop 1002 name")
}

func (u *UpdateTs) TestSelectThenUpdateOtoORelation() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	fillDb(u.Assert(), dbAdapter)
	session := d3Orm.CreateSession()

	repo, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1001"))
	u.NoError(err)

	shop1i.(*Shop).Profile.Unwrap().(*ShopProfile).Description = "new shop 1001 profile"

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM profile_p WHERE description = $1", "new shop 1001 profile")
}

func (u *UpdateTs) TestSelectThenDeleteOtoORelation() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	fillDb(u.Assert(), dbAdapter)
	session := d3Orm.CreateSession()

	repo, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1001"))
	u.NoError(err)

	shop1i.(*Shop).Profile = entity.NewWrapEntity(nil)

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE profile_id IS NULL")
}

func (u *UpdateTs) TestSelectThenViewButDontChangeOtoORelation() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	fillDb(u.Assert(), dbAdapter)
	session := d3Orm.CreateSession()

	repo, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1001"))
	u.NoError(err)

	// previous description and new are equal, we expect 0 updates
	shop1i.(*Shop).Profile.Unwrap().(*ShopProfile).Description = "desc1"

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(0, dbAdapter.UpdateCounter())
}

func (u *UpdateTs) TestSelectThenFullUpdate() {
	dbAdapter := helpers.NewDbAdapterWithQueryCounter(adapter.NewGoPgXAdapter(u.pgDb, &adapter.SquirrelAdapter{}))
	d3Orm := orm.NewOrm(dbAdapter)
	u.NoError(d3Orm.Register((*Book)(nil), (*Shop)(nil), (*ShopProfile)(nil), (*Author)(nil)))

	fillDb(u.Assert(), dbAdapter)
	session := d3Orm.CreateSession()

	repo, err := d3Orm.CreateRepository(session, (*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(repo.CreateQuery().AndWhere("shop_p.id = 1001"))
	u.NoError(err)

	shop1 := shop1i.(*Shop)

	newProfile := &ShopProfile{Description: "new shop profile"}
	shop1.Profile = entity.NewWrapEntity(newProfile)
	shop1.Name = "new shop"

	shop1.Books.Remove(0)

	newAuthor := &Author{Name: "new author"}
	newBook := &Book{Name: "new book", Authors: entity.NewCollection([]interface{}{newAuthor})}
	shop1.Books.Add(newBook)

	oldBook := shop1.Books.Get(0).(*Book)
	oldBook.Authors.Remove(0)

	dbAdapter.ResetCounters()
	u.NoError(session.Flush())

	u.Equal(1, dbAdapter.DeleteCounter())
	u.Equal(2, dbAdapter.UpdateCounter())
	u.Equal(4, dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1 and profile_id = $2", "new shop", newProfile.Id).
		SeeOne("SELECT * FROM book_p WHERE id = $1", newBook.Id).
		SeeOne("SELECT * FROM author_p WHERE id = $1", newAuthor.Id).
		SeeOne("SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", newBook.Id, newAuthor.Id)
}

func fillDb(assert *assert.Assertions, s orm.Storage) {
	err := s.Insert("shop_p", []string{"id", "name", "profile_id"}, nil, []interface{}{1001, "shop1", 1001}, false, nil)
	assert.NoError(err)
	err = s.Insert("shop_p", []string{"id", "name", "profile_id"}, nil, []interface{}{1002, "shop2", 1002}, false, nil)
	assert.NoError(err)

	err = s.Insert("profile_p", []string{"id", "description"}, nil, []interface{}{1001, "desc1"}, false, nil)
	assert.NoError(err)
	err = s.Insert("profile_p", []string{"id", "description"}, nil, []interface{}{1002, "desc2"}, false, nil)
	assert.NoError(err)

	err = s.Insert("book_p", []string{"id", "shop_id", "name"}, nil, []interface{}{1001, 1001, "book1"}, false, nil)
	assert.NoError(err)
	err = s.Insert("book_p", []string{"id", "shop_id", "name"}, nil, []interface{}{1002, 1001, "desc2"}, false, nil)
	assert.NoError(err)
	err = s.Insert("book_p", []string{"id", "shop_id", "name"}, nil, []interface{}{1003, 1002, "desc3"}, false, nil)
	assert.NoError(err)

	err = s.Insert("author_p", []string{"id", "name"}, nil, []interface{}{1001, "author1"}, false, nil)
	assert.NoError(err)
	err = s.Insert("author_p", []string{"id", "name"}, nil, []interface{}{1002, "author2"}, false, nil)
	assert.NoError(err)

	err = s.Insert("book_author_p", []string{"book_id", "author_id"}, nil, []interface{}{1001, 1001}, false, nil)
	assert.NoError(err)
	err = s.Insert("book_author_p", []string{"book_id", "author_id"}, nil, []interface{}{1002, 1001}, false, nil)
	assert.NoError(err)
	err = s.Insert("book_author_p", []string{"book_id", "author_id"}, nil, []interface{}{1002, 1002}, false, nil)
	assert.NoError(err)
}