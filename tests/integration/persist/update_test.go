package persist

import (
	"context"
	d3pgx "github.com/godzie44/d3/adapter/pgx"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/tests/helpers"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type UpdateTs struct {
	suite.Suite
	pgDb      *pgx.Conn
	dbAdapter *helpers.DbAdapterWithQueryCounter
	d3Orm     *orm.Orm
	ctx       context.Context
}

func (u *UpdateTs) SetupSuite() {
	cfg, _ := pgx.ParseConfig(os.Getenv("D3_PG_TEST_DB"))
	driver, err := d3pgx.NewPgxDriver(cfg)
	u.NoError(err)
	u.pgDb, _ = driver.UnwrapConn().(*pgx.Conn)

	err = createSchema(u.pgDb)
	u.NoError(err)

	u.dbAdapter = helpers.NewDbAdapterWithQueryCounter(driver)
	u.d3Orm = orm.New(u.dbAdapter)
	u.NoError(u.d3Orm.Register(
		(*Book)(nil),
		(*Shop)(nil),
		(*ShopProfile)(nil),
		(*Author)(nil),
	))

	u.NoError(err)
}

func (u *UpdateTs) SetupTest() {
	u.ctx = u.d3Orm.CtxWithSession(context.Background())
}

func (u *UpdateTs) TearDownSuite() {
	u.NoError(deleteSchema(u.pgDb))
}

func (u *UpdateTs) TearDownTest() {
	u.NoError(clearSchema(u.pgDb))
	u.dbAdapter.ResetCounters()
}

func TestUpdateSuite(t *testing.T) {
	suite.Run(t, new(UpdateTs))
}

func (u *UpdateTs) TestInsertThenUpdate() {
	shop, err := createAndPersistsShop(u.ctx, u.d3Orm)
	u.NoError(err)

	u.NoError(orm.Session(u.ctx).Flush())

	shop.Name = "new shop"

	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name='new shop'").
		See(0, "SELECT * FROM shop_p WHERE name='shop'")
}

func (u *UpdateTs) TestInsertThenUpdateOToMRelation() {
	shop, err := createAndPersistsShop(u.ctx, u.d3Orm)
	u.NoError(err)

	u.NoError(orm.Session(u.ctx).Flush())

	shop.Books.Remove(0)

	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NOT NULL").
		SeeOne("SELECT * FROM book_p WHERE shop_id IS NULL")
}

func (u *UpdateTs) TestInsertThenUpdateMToMRelations() {
	shop, err := createAndPersistsShop(u.ctx, u.d3Orm)
	u.NoError(err)

	u.NoError(orm.Session(u.ctx).Flush())

	book := shop.Books.Get(0).(*Book)
	author := book.Authors.Get(1).(*Author)

	book.Authors.Remove(1)

	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.DeleteCounter())
	helpers.NewPgTester(u.T(), u.pgDb).
		See(0, "SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", book.Id, author.Id)
}

func (u *UpdateTs) TestInsertThenFullUpdate() {
	shop, err := createAndPersistsShop(u.ctx, u.d3Orm)
	u.NoError(err)

	u.NoError(orm.Session(u.ctx).Flush())

	newProfile := &ShopProfile{Description: "new shop profile"}
	shop.Profile = entity.NewCell(newProfile)
	shop.Name = "new shop"

	shop.Books.Remove(0)

	newAuthor := &Author{Name: "new author"}
	newBook := &Book{Name: "new book", Authors: entity.NewCollection(newAuthor)}
	shop.Books.Add(newBook)

	oldBook := shop.Books.Get(0).(*Book)
	oldBook.Authors.Remove(1)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.DeleteCounter())
	u.Equal(2, u.dbAdapter.UpdateCounter())
	u.Equal(4, u.dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1 and profile_id = $2", "new shop", newProfile.Id).
		SeeOne("SELECT * FROM book_p WHERE id = $1", newBook.Id).
		SeeOne("SELECT * FROM author_p WHERE id = $1", newAuthor.Id).
		SeeOne("SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", newBook.Id, newAuthor.Id)
}

func createAndPersistsShop(ctx context.Context, orm *orm.Orm) (*Shop, error) {
	repository, _ := orm.MakeRepository((*Shop)(nil))

	author1 := &Author{
		Name: "author1",
	}
	shop := &Shop{
		Books: entity.NewCollection(&Book{
			Authors: entity.NewCollection(author1, &Author{Name: "author 2"}),
			Name:    "book 1",
		}, &Book{
			Authors: entity.NewCollection(author1, &Author{Name: "author 3"}),
			Name:    "book 2",
		}),
		Profile: entity.NewCell(&ShopProfile{
			Description: "this is tests shop",
		}),
		Name: "shop",
	}

	return shop, repository.Persists(ctx, shop)
}

func (u *UpdateTs) TestSelectThenSimpleUpdate() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)
	shop2i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1002))
	u.NoError(err)

	shop1i.(*Shop).Name = "new shop 1001 name"
	shop2i.(*Shop).Name = "new shop 1002 name"

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(2, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1", "new shop 1001 name").
		SeeOne("SELECT * FROM shop_p WHERE name = $1", "new shop 1002 name")
}

func (u *UpdateTs) TestSelectThenUpdateOtoORelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	shop1i.(*Shop).Profile.Unwrap().(*ShopProfile).Description = "new shop 1001 profile"

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM profile_p WHERE description = $1", "new shop 1001 profile")
}

func (u *UpdateTs) TestSelectThenDeleteOtoORelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	shop1i.(*Shop).Profile = entity.NewCell(nil)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE profile_id IS NULL")
}

func (u *UpdateTs) TestSelectThenChangeOtoORelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	shop1i.(*Shop).Profile = entity.NewCell(&ShopProfile{
		Description: "changed profile",
	})

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE id = 1001 AND profile_id IS NOT NULL").
		SeeOne("SELECT * FROM profile_p WHERE description='changed profile'")
}

func (u *UpdateTs) TestSelectThenViewButDontChangeOtoORelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	// previous description and new are equal, we expect 0 updates
	shop1i.(*Shop).Profile.Unwrap().(*ShopProfile).Description = "desc1"

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(0, u.dbAdapter.UpdateCounter())
}

func (u *UpdateTs) TestSelectThenUpdateOtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	shop1.(*Shop).Books.Get(0).(*Book).Name = "new book 0"
	shop1.(*Shop).Books.Get(1).(*Book).Name = "new book 1"

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(2, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM book_p WHERE name = $1", "new book 0").
		SeeOne("SELECT * FROM book_p WHERE name = $1", "new book 1")
}

func (u *UpdateTs) TestSelectThenDeleteOtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	oldBookCount := shop1.(*Shop).Books.Count()
	shop1.(*Shop).Books.Remove(0)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		See(oldBookCount-1, "SELECT * FROM book_p WHERE shop_id = 1001")
}

func (u *UpdateTs) TestSelectThenAddOtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	newBook := &Book{
		Authors: nil,
		Name:    "new book",
	}
	shop1.(*Shop).Books.Add(newBook)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM book_p WHERE shop_id = 1001 AND name = 'new book'")
}

func (u *UpdateTs) TestSelectThenViewButDontChangeOtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	sameName := shop1.(*Shop).Books.Get(0).(*Book).Name
	shop1.(*Shop).Books.Get(0).(*Book).Name = sameName

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(0, u.dbAdapter.UpdateCounter())
}

func (u *UpdateTs) TestSelectThenUpdateMtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Book)(nil))
	u.NoError(err)

	book1, err := repo.FindOne(u.ctx, repo.Select().Where("book_p.id", "=", 1002))
	u.NoError(err)

	book1.(*Book).Authors.Get(0).(*Author).Name = "new author 1"
	book1.(*Book).Authors.Get(1).(*Author).Name = "new author 2"

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(2, u.dbAdapter.UpdateCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM author_p WHERE name = $1", "new author 1").
		SeeOne("SELECT * FROM author_p WHERE name = $1", "new author 2")
}

func (u *UpdateTs) TestSelectThenDeleteMtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Book)(nil))
	u.NoError(err)

	book1, err := repo.FindOne(u.ctx, repo.Select().Where("book_p.id", "=", 1002))
	u.NoError(err)

	oldAuthorCount := book1.(*Book).Authors.Count()
	book1.(*Book).Authors.Remove(1)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.DeleteCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		See(oldAuthorCount-1, "SELECT * FROM book_author_p WHERE book_id = 1002")
}

func (u *UpdateTs) TestSelectThenAddMtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Book)(nil))
	u.NoError(err)

	book1, err := repo.FindOne(u.ctx, repo.Select().Where("book_p.id", "=", 1001))
	u.NoError(err)
	book2, err := repo.FindOne(u.ctx, repo.Select().Where("book_p.id", "=", 1002))
	u.NoError(err)

	newAuthor := &Author{
		Name: "new author",
	}
	book1.(*Book).Authors.Add(newAuthor)
	book2.(*Book).Authors.Add(newAuthor)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(3, u.dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM author_p WHERE name = 'new author'").
		SeeOne("SELECT * FROM book_author_p WHERE book_id = 1001 AND author_id = $1", newAuthor.Id).
		SeeOne("SELECT * FROM book_author_p WHERE book_id = 1002 AND author_id = $1", newAuthor.Id)
}

func (u *UpdateTs) TestSelectThenViewButDontChangeMtoMRelation() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Book)(nil))
	u.NoError(err)

	book1, err := repo.FindOne(u.ctx, repo.Select().Where("book_p.id", "=", 1002))
	u.NoError(err)

	sameName := book1.(*Book).Authors.Get(0).(*Author).Name
	book1.(*Book).Authors.Get(0).(*Author).Name = sameName

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(0, u.dbAdapter.UpdateCounter())
}

func (u *UpdateTs) TestSelectThenFullUpdate() {
	fillDb(u.Assert(), u.dbAdapter)

	repo, err := u.d3Orm.MakeRepository((*Shop)(nil))
	u.NoError(err)

	shop1i, err := repo.FindOne(u.ctx, repo.Select().Where("shop_p.id", "=", 1001))
	u.NoError(err)

	shop1 := shop1i.(*Shop)

	newProfile := &ShopProfile{Description: "new shop profile"}
	shop1.Profile = entity.NewCell(newProfile)
	shop1.Name = "new shop"

	shop1.Books.Remove(0)

	newAuthor := &Author{Name: "new author"}
	newBook := &Book{Name: "new book", Authors: entity.NewCollection(newAuthor)}
	shop1.Books.Add(newBook)

	oldBook := shop1.Books.Get(0).(*Book)
	oldBook.Authors.Remove(0)

	u.dbAdapter.ResetCounters()
	u.NoError(orm.Session(u.ctx).Flush())

	u.Equal(1, u.dbAdapter.DeleteCounter())
	u.Equal(2, u.dbAdapter.UpdateCounter())
	u.Equal(4, u.dbAdapter.InsertCounter())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM shop_p WHERE name = $1 and profile_id = $2", "new shop", newProfile.Id).
		SeeOne("SELECT * FROM book_p WHERE id = $1", newBook.Id).
		SeeOne("SELECT * FROM author_p WHERE id = $1", newAuthor.Id).
		SeeOne("SELECT * FROM book_author_p WHERE book_id = $1 and author_id = $2", newBook.Id, newAuthor.Id)
}

func (u *UpdateTs) TestDoubleAddMToMIsIdempotence() {
	repo, err := u.d3Orm.MakeRepository((*Book)(nil))
	u.NoError(err)

	author := &Author{Name: "author"}
	book := &Book{
		Authors: entity.NewCollection(author),
		Name:    "book",
	}

	u.NoError(repo.Persists(u.ctx, book))
	u.NoError(orm.Session(u.ctx).Flush())

	newCtx := u.d3Orm.CtxWithSession(context.Background())
	fetchedEntity, err := repo.FindOne(newCtx, repo.Select().Where("id", "=", book.Id))
	u.NoError(err)

	fetchedBook := fetchedEntity.(*Book)
	fetchedBook.Authors.Add(author)

	u.NoError(repo.Persists(newCtx, book))
	u.NoError(orm.Session(newCtx).Flush())

	helpers.NewPgTester(u.T(), u.pgDb).
		SeeOne("SELECT * FROM book_author_p")
}
