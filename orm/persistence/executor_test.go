package persistence

import (
	"d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type scannerStub struct {
	ret []interface{}
}

func (s *scannerStub) Scan(dst ...interface{}) error {
	for i, ptr := range dst {
		val := reflect.ValueOf(ptr)
		val.Elem().Set(reflect.ValueOf(s.ret[i]))
	}

	return nil
}

type pusherStub struct {
	insertRequests []queryStub
}

type queryStub struct {
	tableName string
	values    map[string]interface{}
}

func (s *pusherStub) Insert(table string, cols []string, values []interface{}) error {
	qValues := map[string]interface{}{}
	for i, val := range values {
		if fn, ok := val.(func() (interface{}, error)); ok {
			fnVal, _ := fn()
			qValues[cols[i]] = fnVal
		} else {
			qValues[cols[i]] = val
		}
	}

	s.insertRequests = append(s.insertRequests, queryStub{
		tableName: table,
		values:    qValues,
	})

	return nil
}

func (s *pusherStub) InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner Scanner) error) error {
	qValues := map[string]interface{}{}
	for i, val := range values {
		if fn, ok := val.(func() (interface{}, error)); ok {
			fnVal, _ := fn()
			qValues[cols[i]] = fnVal
		} else {
			qValues[cols[i]] = val
		}
	}

	s.insertRequests = append(s.insertRequests, queryStub{
		tableName: table,
		values:    qValues,
	})

	return withReturned(&scannerStub{ret: []interface{}{values[0]}})
}

func (s *pusherStub) Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error {
	return nil
}

func (s *pusherStub) Remove(_ string, _ map[string]interface{}) error {
	return nil
}

//assertQueryAfter check that query will be execute after queryBefore
func (s *pusherStub) assertQueryAfter(t *testing.T, query, queryBefore queryStub) {
	var queryIndex, queryBeforeIndex int = -1, -1
	for i := range s.insertRequests {
		if reflect.DeepEqual(s.insertRequests[i], query) {
			queryIndex = i
		} else if reflect.DeepEqual(s.insertRequests[i], queryBefore) {
			queryBeforeIndex = i
		}
	}

	if queryIndex == -1 {
		assert.Fail(t, "query not found")
	}
	if queryBeforeIndex == -1 {
		assert.Fail(t, "query before not found")
	}

	assert.True(t, queryIndex > queryBeforeIndex)
}

func TestExecuteSimpleGraph(t *testing.T) {
	testPusher := &pusherStub{}
	executor := NewExecutor(testPusher, func(act CompositeAction) {})

	meta, _ := metaRegistry.GetMeta((*Shop)(nil))
	shop := &Shop{ID: 1, Profile: entity.NewWrapEntity(&ShopProfile{ID: 1})}

	graph := createNewGraph()

	_ = graph.ProcessEntity(entity.NewBox(shop, &meta))

	err := executor.Exec(graph)
	assert.NoError(t, err)

	testPusher.assertQueryAfter(t, queryStub{
		tableName: "shop",
		values:    map[string]interface{}{"id": 1, "profile_id": 1},
	}, queryStub{
		tableName: "profile",
		values:    map[string]interface{}{"id": 1},
	})
}

func TestExecuteComplexGraph(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Shop)(nil))

	commonAuthor := &Author{
		ID: 1,
	}
	authors := []interface{}{
		commonAuthor,
		&Author{
			ID: 2,
		},
	}

	books := []interface{}{
		&Book{
			ID:      1,
			Authors: entity.NewCollection([]interface{}{commonAuthor}),
		},
		&Book{
			ID:      3,
			Authors: entity.NewCollection(authors),
		},
	}

	shop1 := &Shop{
		ID:      1,
		Profile: entity.NewWrapEntity(&ShopProfile{ID: 1}),
		Books:   entity.NewCollection(books)}
	shop2 := &Shop{
		ID:      2,
		Profile: entity.NewWrapEntity(&ShopProfile{ID: 2}),
		Books: entity.NewCollection([]interface{}{&Book{
			ID: 2,
		}}),
	}

	graph := createNewGraph()
	_ = graph.ProcessEntity(entity.NewBox(shop1, &meta))
	_ = graph.ProcessEntity(entity.NewBox(shop2, &meta))

	testStorage := &pusherStub{}
	executor := NewExecutor(testStorage, func(act CompositeAction) {})
	err := executor.Exec(graph)
	assert.NoError(t, err)

	testStorage.assertQueryAfter(t, queryStub{tableName: "shop", values: map[string]interface{}{"id": 1, "profile_id": 1}},
		queryStub{tableName: "profile", values: map[string]interface{}{"id": 1}})

	testStorage.assertQueryAfter(t, queryStub{tableName: "book", values: map[string]interface{}{"id": 1, "shop_id": 1}},
		queryStub{tableName: "shop", values: map[string]interface{}{"id": 1, "profile_id": 1}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book", values: map[string]interface{}{"id": 3, "shop_id": 1}},
		queryStub{tableName: "shop", values: map[string]interface{}{"id": 1, "profile_id": 1}})

	testStorage.assertQueryAfter(t, queryStub{tableName: "book_author", values: map[string]interface{}{"book_id": 3, "author_id": 1}},
		queryStub{tableName: "book", values: map[string]interface{}{"id": 3, "shop_id": 1}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book_author", values: map[string]interface{}{"book_id": 3, "author_id": 2}},
		queryStub{tableName: "book", values: map[string]interface{}{"id": 3, "shop_id": 1}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book_author", values: map[string]interface{}{"book_id": 3, "author_id": 1}},
		queryStub{tableName: "author", values: map[string]interface{}{"id": 1}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book_author", values: map[string]interface{}{"book_id": 3, "author_id": 2}},
		queryStub{tableName: "author", values: map[string]interface{}{"id": 2}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book_author", values: map[string]interface{}{"book_id": 1, "author_id": 1}},
		queryStub{tableName: "book", values: map[string]interface{}{"id": 1, "shop_id": 1}})

	testStorage.assertQueryAfter(t, queryStub{tableName: "shop", values: map[string]interface{}{"id": 2, "profile_id": 2}}, queryStub{tableName: "profile", values: map[string]interface{}{"id": 2}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book", values: map[string]interface{}{"id": 2, "shop_id": 2}},
		queryStub{tableName: "shop", values: map[string]interface{}{"id": 2, "profile_id": 2}})
}

type Order struct {
	entity   struct{}             `d3:"table_name:order"`
	Id       int                  `d3:"pk:auto"`
	Items    entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/persistence/OrderItem,join_on:order_id>,type:lazy"`
	BestItem entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/OrderItem,join_on:best_item_id>,type:lazy"`
}

type OrderItem struct {
	entity struct{} `d3:"table_name:order_item"`
	Id     int      `d3:"pk:auto"`
}

func TestExecuteWithCircularReference(t *testing.T) {
	metaRegistry.Add((*Order)(nil), (*OrderItem)(nil))

	bestItem := &OrderItem{Id: 1}
	orderItems := []interface{}{
		bestItem,
		&OrderItem{Id: 2},
	}

	order := &Order{
		Id:       1,
		Items:    entity.NewCollection(orderItems),
		BestItem: entity.NewWrapEntity(bestItem),
	}

	meta, _ := metaRegistry.GetMeta(order)
	graph := createNewGraph()
	_ = graph.ProcessEntity(entity.NewBox(order, &meta))

	r := hasCycle(graph)
	assert.False(t, r)
	//nodes := graph.filterRoots()
	//testStorage := &pusherStub{}
	//err := NewExecutor(testStorage).exec(graph)
	//assert.NoError(t, err)
}
