package benchmark

import (
	"d3/orm"
	"d3/orm/entity"
	"d3/orm/persistence"
	"d3/orm/query"
	"reflect"
)

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

func (i *inMemoryStorage) MakeRawDataMapper() orm.RawDataMapper {
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

func (p *pusherStub) Insert(table string, cols []string, values []interface{}) error {
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
	if err := p.Insert(table, append(cols, returnCols...), append(values, newId)); err != nil {
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