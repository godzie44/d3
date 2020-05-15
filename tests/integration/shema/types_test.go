package shema

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"database/sql"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type allTypeStruct struct {
	ID               sql.NullInt32 `d3:"pk:auto"`
	BoolField        bool
	IntField         int
	Int32Field       int32
	Int64Field       int64
	Float32Field     float32
	Float64Field     float64
	StringField      string
	TimeField        time.Time
	NullBoolField    sql.NullBool
	NullI32Field     sql.NullInt32
	NullI64Field     sql.NullInt64
	NullFloat64Field sql.NullFloat64
	NullStringField  sql.NullString
	NullTimeField    sql.NullTime
}

func TestTypeConversion(t *testing.T) {
	pgDb, d3orm := initDb(t)

	s1 := d3orm.MakeSession()
	rep, err := s1.MakeRepository(&allTypeStruct{})
	assert.NoError(t, err)

	currTime := time.Unix(time.Now().Unix(), 0)
	entity := &allTypeStruct{
		BoolField:        true,
		IntField:         1,
		Int32Field:       2,
		Int64Field:       3,
		Float32Field:     4,
		Float64Field:     5,
		StringField:      "d3",
		TimeField:        currTime,
		NullBoolField:    sql.NullBool{Bool: true, Valid: true},
		NullI32Field:     sql.NullInt32{Int32: 1, Valid: true},
		NullI64Field:     sql.NullInt64{Int64: 2, Valid: true},
		NullFloat64Field: sql.NullFloat64{Float64: 1.1, Valid: true},
		NullStringField:  sql.NullString{Valid: false},
		NullTimeField:    sql.NullTime{Time: currTime, Valid: true},
	}

	assert.NoError(t, rep.Persists(entity))
	assert.NoError(t, s1.Flush())

	s2 := d3orm.MakeSession()
	rep, err = s2.MakeRepository(&allTypeStruct{})
	assert.NoError(t, err)

	fetchedEntity, err := rep.FindOne(rep.CreateQuery().AndWhere("id = ?", entity.ID))
	assert.NoError(t, err)

	assert.Equal(t, entity, fetchedEntity)
	dropDb(t, pgDb)
}

func initDb(t *testing.T) (*pgx.Conn, *orm.Orm) {
	pgDb, _ := pgx.Connect(context.Background(), os.Getenv("D3_PG_TEST_DB"))

	d3orm := orm.NewOrm(adapter.NewGoPgXAdapter(pgDb, &adapter.SquirrelAdapter{}))
	assert.NoError(t, d3orm.Register(orm.NewMapping("all_types", (*allTypeStruct)(nil))))

	sql, err := d3orm.GenerateSchema()
	assert.NoError(t, err)

	_, err = pgDb.Exec(context.Background(), sql)
	assert.NoError(t, err)

	return pgDb, d3orm
}

func dropDb(t *testing.T, db *pgx.Conn) {
	_, err := db.Exec(context.Background(), `
DROP TABLE IF EXISTS all_types;
`)
	assert.NoError(t, err)
}
