package persist

import (
	"context"
	"d3/orm"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func createSchema(db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS shop_p(
		id SERIAL PRIMARY KEY,
		profile_id integer,
		name character varying(200) NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS profile_p(
		id SERIAL PRIMARY KEY,
		description character varying(1000) NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book_p(
		id SERIAL PRIMARY KEY,
		name text NOT NULL,
		shop_id integer
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS author_p(
		id SERIAL PRIMARY KEY,
		name character varying(200) NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS book_author_p(
		book_id integer NOT NULL,
		author_id integer NOT NULL
	)`)
	return err
}

func deleteSchema(db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), `
DROP TABLE book_p;
DROP TABLE author_p;
DROP TABLE book_author_p;
DROP TABLE shop_p;
DROP TABLE profile_p;
`)
	return err
}

func clearSchema(db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), `
delete from book_p;
delete from author_p;
delete from book_author_p;
delete from shop_p;
delete from profile_p;
`)
	return err
}

func fillDb(assert *assert.Assertions, s orm.Storage) {
	tx, err := s.BeginTx()
	assert.NoError(err)

	ps := s.MakePusher(tx)
	err = ps.Insert("shop_p", []string{"id", "name", "profile_id"}, []interface{}{1001, "shop1", 1001})
	assert.NoError(err)
	err = ps.Insert("shop_p", []string{"id", "name", "profile_id"}, []interface{}{1002, "shop2", 1002})
	assert.NoError(err)

	err = ps.Insert("profile_p", []string{"id", "description"}, []interface{}{1001, "desc1"})
	assert.NoError(err)
	err = ps.Insert("profile_p", []string{"id", "description"}, []interface{}{1002, "desc2"})
	assert.NoError(err)

	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1001, 1001, "book1"})
	assert.NoError(err)
	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1002, 1001, "book2"})
	assert.NoError(err)
	err = ps.Insert("book_p", []string{"id", "shop_id", "name"}, []interface{}{1003, 1002, "book3"})
	assert.NoError(err)

	err = ps.Insert("author_p", []string{"id", "name"}, []interface{}{1001, "author1"})
	assert.NoError(err)
	err = ps.Insert("author_p", []string{"id", "name"}, []interface{}{1002, "author2"})
	assert.NoError(err)

	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1001, 1001})
	assert.NoError(err)
	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1002, 1001})
	assert.NoError(err)
	err = ps.Insert("book_author_p", []string{"book_id", "author_id"}, []interface{}{1002, 1002})
	assert.NoError(err)

	assert.NoError(tx.Commit())
}
