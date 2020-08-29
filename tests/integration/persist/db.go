package persist

import (
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/stretchr/testify/assert"
)

func fillDb(assert *assert.Assertions, s orm.Driver) {
	tx, err := s.BeginTx()
	assert.NoError(err)

	ps := s.MakePusher(tx)
	err = ps.Insert("shop_p", []string{"id", "name", "profile_id"}, []interface{}{1001, "shop1", 1001}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("shop_p", []string{"id", "name", "profile_id"}, []interface{}{1002, "shop2", 1002}, persistence.Undefined)
	assert.NoError(err)

	err = ps.Insert("profile_p", []string{"id", "description"}, []interface{}{1001, "desc1"}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("profile_p", []string{"id", "description"}, []interface{}{1002, "desc2"}, persistence.Undefined)
	assert.NoError(err)

	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1001, 1001, "book1"}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1002, 1001, "book2"}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1003, 1002, "book3"}, persistence.Undefined)
	assert.NoError(err)

	err = ps.Insert("author_p", []string{"id", "name"}, []interface{}{1001, "author1"}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("author_p", []string{"id", "name"}, []interface{}{1002, "author2"}, persistence.Undefined)
	assert.NoError(err)

	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1001, 1001}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1002, 1001}, persistence.Undefined)
	assert.NoError(err)
	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1002, 1002}, persistence.Undefined)
	assert.NoError(err)

	assert.NoError(tx.Commit())
}
