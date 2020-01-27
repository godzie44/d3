package repository

import (
	"context"
	"d3/adapter"
	"d3/orm"
	"d3/orm/entity"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testEntity1WithLazyLoad struct {
	entity struct{}             `d3:"table_name:test_entity_t1"`
	Id     int64                `pg:",pk"`
	Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/repository/testEntity2,join_on:t2_id>,type:lazy"`
	Data   string
}

type testEntity2 struct {
	entity struct{} `d3:"table_name:test_entity_t2"`
	Id     int64    `pg:",pk"`
	Data   string
	//TestEntity1Collection mapper.Collection `d3:"many_to_one"`
}

func createOrm(t *testing.T) *orm.Orm {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "postgres", "0.0.0.0:5432", "repr")
	pgDb, err := pgx.Connect(context.Background(), dsn)
	assert.Nil(t, err)

	return orm.NewOrm(adapter.NewGoPgXAdapter(pgDb, &adapter.SquirrelAdapter{}))
}

//func TestRepositoryFindOneEntityWithLazyRelation(t *testing.T) {
//	stormOrm := createOrm(t)
//
//	err := stormOrm.Register((*testEntity1WithLazyLoad)(nil), (*testEntity2)(nil))
//	assert.Nil(t, err)
//
//	session := stormOrm.CreateSession()
//	repository, err := stormOrm.CreateRepository(session, (*testEntity1WithLazyLoad)(nil))
//	assert.Nil(t, err)
//
//	entity, err := repository.FindOne(query.NewQuery().AndWhere("id = ?", 1))
//	assert.Nil(t, err)
//
//	fmt.Println(entity)
//	fmt.Println(entity.(*testEntity1WithLazyLoad).Rel.Unwrap())
//
//	assert.Nil(t, err)
//	assert.IsType(t, &testEntity1WithLazyLoad{}, entity)
//	assert.Equal(t, []interface{}{int64(1), "data"}, []interface{}{entity.(*testEntity1WithLazyLoad).Id, entity.(*testEntity1WithLazyLoad).Data})
//
//	relatedEntity := entity.(*testEntity1WithLazyLoad).Rel.Unwrap().(*testEntity2)
//	assert.IsType(t, &testEntity2{}, relatedEntity)
//	assert.Equal(t, []interface{}{int64(1), "haha"}, []interface{}{relatedEntity.Id, relatedEntity.Data})
//}
//
//func TestRepositoryFindOneEntity(t *testing.T) {
//	orm := createOrm(t)
//	_ = orm.Register((*testEntity1WithLazyLoad)(nil), (*testEntity2)(nil))
//
//	session := orm.CreateSession()
//	repository, _ := orm.CreateRepository(session, (*testEntity1WithLazyLoad)(nil))
//	entities, err := repository.FindAll(query.NewQuery())
//
//	assert.Nil(t, err)
//	assert.IsType(t, []*testEntity1WithLazyLoad{}, entities)
//	assert.Len(t, entities, 2)
//}
//
//type testEntity1WithEagerLoad struct {
//	entity struct{}             `d3:"table_name:test_entity_t1"`
//	Id     int64                `pg:",pk"`
//	Rel    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/test/integration/repository/testEntity2,join_on:t2_id,reference_on:id>,type:eager"`
//	Data   string
//}
//
//func TestRepositoryFindOneEntityWithEagerRelation(t *testing.T) {
//	orm := createOrm(t)
//	_ = orm.Register((*testEntity1WithEagerLoad)(nil), (*testEntity2)(nil))
//
//	session := orm.CreateSession()
//	repository, _ := orm.CreateRepository(session, (*testEntity1WithEagerLoad)(nil))
//	e, err := repository.FindOne(query.NewQuery().AndWhere("test_entity_t1.id = ?", 1))
//	assert.Nil(t, err)
//
//	fmt.Println(e)
//	fmt.Println(e.(*testEntity1WithEagerLoad).Rel.Unwrap())
//
//	assert.Nil(t, err)
//	assert.IsType(t, &testEntity1WithEagerLoad{}, e)
//	assert.IsType(t, &testEntity2{}, e.(*testEntity1WithEagerLoad).Rel.Unwrap())
//}
