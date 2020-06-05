package benchmark

import (
	"d3/orm"
	"d3/orm/entity"
	"runtime"
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	d3orm := orm.NewOrm(newInMemoryStorage())
	_ = d3orm.Register(
		&shop{},
		&profile{},
		&book{},
		&author{},
	)
	sess := d3orm.MakeSession()
	repo, _ := sess.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		aggregate := createAggregate()
		_ = repo.Persists(aggregate)
		_ = sess.Flush()
	}
}

func BenchmarkSelect(b *testing.B) {
	d3orm := orm.NewOrm(newInMemoryStorage())
	_ = d3orm.Register(
		&shop{},
		&profile{},
		&book{},
		&author{},
	)
	sess := d3orm.MakeSession()
	repo, _ := sess.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		res, _ := repo.FindOne(repo.CreateQuery().AndWhere("id = ?", i))
		runtime.KeepAlive(res)
	}
}

func BenchmarkUpdate(b *testing.B) {
	d3orm := orm.NewOrm(newInMemoryStorage())
	_ = d3orm.Register(
		&shop{},
		&profile{},
		&book{},
		&author{},
	)
	sess := d3orm.MakeSession()
	repo, _ := sess.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		res, _ := repo.FindOne(repo.CreateQuery().AndWhere("id = ?", i))

		shop := res.(*shop)
		shop.name += " updated"

		book := shop.books.Get(0).(*book)
		book.Name += " updated"
		_ = sess.Flush()
	}
}

func createAggregate() *shop {
	author1 := &author{
		Name: "a1",
	}
	author2 := &author{
		Name: "a2",
	}

	book := &book{
		Authors: entity.NewCollection(author1, author2),
		Name:    "new book",
	}

	shop := &shop{
		books: entity.NewCollection(book),
		profile: entity.NewWrapEntity(&profile{
			Description: "good shop",
		}),
		name: "new shop",
	}

	return shop
}
