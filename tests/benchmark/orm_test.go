package benchmark

import (
	"context"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
	"reflect"
	"runtime"
	"testing"
)

func createOrm() *orm.Orm {
	d3orm := orm.New(newInMemoryStorage())
	_ = d3orm.Register(
		&shop{},
		&profile{},
		&book{},
		&author{},
	)
	return d3orm
}

func BenchmarkInsert(b *testing.B) {
	d3orm := createOrm()

	ctx := d3orm.CtxWithSession(context.Background())
	repo, _ := d3orm.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		aggregate := createAggregate()
		_ = repo.Persists(ctx, aggregate)
		_ = orm.Session(ctx).Flush()
	}
}

func BenchmarkSelect(b *testing.B) {
	d3orm := createOrm()

	ctx := d3orm.CtxWithSession(context.Background())
	repo, _ := d3orm.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		res, _ := repo.FindOne(ctx, repo.MakeQuery().AndWhere("id = ?", i))
		runtime.KeepAlive(res)
	}
}

func BenchmarkUpdate(b *testing.B) {
	d3orm := createOrm()

	ctx := d3orm.CtxWithSession(context.Background())
	repo, _ := d3orm.MakeRepository(&shop{})
	for i := 0; i < b.N; i++ {
		res, _ := repo.FindOne(ctx, repo.MakeQuery().AndWhere("id = ?", i))

		shop := res.(*shop)
		shop.name += " updated"

		book := shop.books.Get(0).(*book)
		book.Name += " updated"
		_ = orm.Session(ctx).Flush()
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
		profile: entity.NewCell(&profile{
			Description: "good shop",
		}),
		name: "new shop",
	}

	return shop
}

type inMemoryStorage struct {
	pusher persistence.Pusher
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{pusher: &pusherStub{
		store:      map[string][]map[string]interface{}{},
		idCounters: map[string]int{},
	}}

}

func (i *inMemoryStorage) MakePusher(_ orm.Transaction) persistence.Pusher {
	return i.pusher
}

func (i *inMemoryStorage) ExecuteQuery(query *query.Query) ([]map[string]interface{}, error) {
	eName := query.OwnerMeta().EntityName
	switch eName {
	case entity.NameFromEntity(&shop{}):
		return []map[string]interface{}{
			{"shop.id": 1, "shop.name": "shop1", "shop.profile_id": 1},
		}, nil
	case entity.NameFromEntity(&profile{}):
		return []map[string]interface{}{
			{"prof.id": 1, "prof.description": "description1"},
		}, nil
	case entity.NameFromEntity(&book{}):
		return []map[string]interface{}{
			{"book.id": 1, "book.name": "book1"},
		}, nil
	case entity.NameFromEntity(&author{}):
		return []map[string]interface{}{
			{"author.id": 1, "author.name": "author 1"},
		}, nil
	}

	return nil, nil
}

func (i *inMemoryStorage) BeforeQuery(_ func(query string, args ...interface{})) {
}

func (i *inMemoryStorage) AfterQuery(_ func(query string, args ...interface{})) {
}

func (i *inMemoryStorage) BeginTx() (orm.Transaction, error) {
	return &txStub{}, nil
}

func (i *inMemoryStorage) MakeScalarDataMapper() orm.ScalarDataMapper {
	return func(data interface{}, _ reflect.Kind) interface{} {
		return data
	}
}

type txStub struct {
}

func (t *txStub) Commit() error {
	return nil
}

func (t *txStub) Rollback() error {
	return nil
}

type pusherStub struct {
	store      map[string][]map[string]interface{}
	idCounters map[string]int
}

func (p *pusherStub) Insert(table string, cols []string, values []interface{}, _ persistence.OnConflict) error {
	colsVals := make(map[string]interface{})
	for i, col := range cols {
		colsVals[col] = values[i]
	}

	p.store[table] = append(p.store[table], colsVals)
	return nil
}

type idScanner struct {
	id int
}

func (i *idScanner) Scan(v ...interface{}) error {
	if len(v) > 0 {
		v[0] = i.id
	}
	return nil
}

func (p *pusherStub) InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner persistence.Scanner) error) error {
	if len(returnCols) > 1 {
		panic("unsupported return col count")
	}

	newId := p.idCounters[table] + 1
	if err := p.Insert(table, append(cols, returnCols...), append(values, newId), persistence.Undefined); err != nil {
		return err
	}

	p.idCounters[table] = newId

	return withReturned(&idScanner{id: newId})
}

func (p *pusherStub) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
	for rowInd, row := range p.store[table] {
		if rowIsIdentified(row, identityCond) {
			for i := range cols {
				p.store[table][rowInd][cols[i]] = values[i]
			}
		}
	}

	return nil
}

func rowIsIdentified(row map[string]interface{}, identityCond map[string]interface{}) bool {
	for col, val := range identityCond {
		if row[col] != val {
			return false
		}
	}
	return true
}

func (p *pusherStub) Remove(table string, identityCond map[string]interface{}) error {
	var delIdx = -1
	for rowIdx, row := range p.store[table] {
		if rowIsIdentified(row, identityCond) {
			delIdx = rowIdx
			break
		}
	}

	if delIdx != -1 {
		p.store[table] = append(p.store[table][:delIdx], p.store[table][delIdx+1:]...)
	}

	return nil
}
