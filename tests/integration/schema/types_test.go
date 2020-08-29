package schema

import (
	"context"
	"database/sql"
	"github.com/godzie44/d3/orm"
	"github.com/godzie44/d3/tests/helpers/db"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Defer func()

func initPGDb(t *testing.T) (*orm.Orm, Defer) {
	_, d3orm, execSqlFn, _ := db.CreatePGTestComponents(t)
	assert.NoError(t, d3orm.Register((*allTypeStruct)(nil), (*entityWithAliases)(nil)))

	sqlSchema, err := d3orm.GenerateSchema()
	assert.NoError(t, err)

	assert.NoError(t, execSqlFn(sqlSchema))

	return d3orm, func() {
		assert.NoError(t, execSqlFn(`
DROP TABLE IF EXISTS all_types;
DROP TABLE IF EXISTS test_aliases;
`))
	}
}

func initSqliteDb(t *testing.T) (*orm.Orm, Defer) {
	_, d3orm, execSqlFn, _ := db.CreateSQLiteTestComponents(t)
	assert.NoError(t, d3orm.Register((*allTypeStruct)(nil), (*entityWithAliases)(nil)))

	sqlSchema, err := d3orm.GenerateSchema()
	assert.NoError(t, err)

	assert.NoError(t, execSqlFn(sqlSchema))

	return d3orm, func() {
		assert.NoError(t, execSqlFn(`
DROP TABLE IF EXISTS all_types;
DROP TABLE IF EXISTS test_aliases;
`))
	}
}

func assertTypeConversion(t *testing.T, d3orm *orm.Orm, timeVal time.Time) {
	ctx := d3orm.CtxWithSession(context.Background())
	rep, err := d3orm.MakeRepository(&allTypeStruct{})
	assert.NoError(t, err)

	uuidVal, _ := uuid.FromString("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	entity := &allTypeStruct{
		Uuid:             uuidVal,
		BoolField:        true,
		IntField:         1,
		Int32Field:       2,
		Int64Field:       3,
		Float32Field:     4,
		Float64Field:     5,
		StringField:      "d3",
		TimeField:        timeVal,
		NullBoolField:    sql.NullBool{Bool: true, Valid: true},
		NullI32Field:     sql.NullInt32{Int32: 1, Valid: true},
		NullI64Field:     sql.NullInt64{Int64: 2, Valid: true},
		NullFloat64Field: sql.NullFloat64{Float64: 1.1, Valid: true},
		NullStringField:  sql.NullString{Valid: false},
		NullTimeField:    sql.NullTime{Time: timeVal, Valid: true},
	}
	assert.NoError(t, rep.Persists(ctx, entity))
	assert.NoError(t, orm.Session(ctx).Flush())

	ctx2 := d3orm.CtxWithSession(context.Background())
	rep, err = d3orm.MakeRepository(&allTypeStruct{})
	assert.NoError(t, err)

	fetchedEntity, err := rep.FindOne(ctx2, rep.Select().Where("id", "=", entity.ID))
	assert.NoError(t, err)

	assert.Equal(t, entity, fetchedEntity)
}

func assertCustomTypeConversion(t *testing.T, d3orm *orm.Orm) {
	ctx := d3orm.CtxWithSession(context.Background())
	rep, err := d3orm.MakeRepository(&entityWithAliases{})
	assert.NoError(t, err)

	entity := &entityWithAliases{
		email:       Email("mail"),
		secretEmail: myEmail("mail"),
	}

	assert.NoError(t, rep.Persists(ctx, entity))
	assert.NoError(t, orm.Session(ctx).Flush())

	ctx2 := d3orm.CtxWithSession(context.Background())
	rep, err = d3orm.MakeRepository(&entityWithAliases{})
	assert.NoError(t, err)

	fetchedEntity, err := rep.FindOne(ctx2, rep.Select().Where("id", "=", entity.ID))
	assert.NoError(t, err)

	assert.Equal(t, entity, fetchedEntity)
}

func TestPGTypeConversion(t *testing.T) {
	d3orm, deferFn := initPGDb(t)
	defer deferFn()

	assertTypeConversion(t, d3orm, time.Unix(time.Now().Unix(), 0))
}

func TestPGCustomTypeConversion(t *testing.T) {
	d3orm, deferFn := initPGDb(t)
	defer deferFn()

	assertCustomTypeConversion(t, d3orm)
}

func TestSqliteTypeConversion(t *testing.T) {
	d3orm, deferFn := initSqliteDb(t)
	defer deferFn()

	assertTypeConversion(t, d3orm, time.Unix(time.Now().Unix(), 0).UTC())
}

func TestSqliteCustomTypeConversion(t *testing.T) {
	d3orm, deferFn := initSqliteDb(t)
	defer deferFn()

	assertCustomTypeConversion(t, d3orm)
}
