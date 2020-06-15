package persistence

import (
	"github.com/godzie44/d3/orm/entity"
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

func (s *pusherStub) Insert(table string, cols []string, values []interface{}, _ OnConflict) error {
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

func (s *pusherStub) InsertWithReturn(table string, cols []string, values []interface{}, _ []string, withReturned func(scanner Scanner) error) error {
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

func (s *pusherStub) Update(_ string, _ []string, _ []interface{}, _ map[string]interface{}) error {
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

	meta, err := metaRegistry.GetMeta((*Shop)(nil))
	assert.NoError(t, err)

	shop := &Shop{ID: 1, Profile: entity.NewCell(&ShopProfile{ID: 1})}

	graph := createNewGraph()

	_ = graph.ProcessEntity(entity.NewBox(shop, &meta))

	err = executor.Exec(graph)
	assert.NoError(t, err)

	testPusher.assertQueryAfter(t, queryStub{
		tableName: "shop",
		values:    map[string]interface{}{"id": 1, "profile_id": 1},
	}, queryStub{
		tableName: "shopprofile",
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
			Authors: entity.NewCollection(commonAuthor),
		},
		&Book{
			ID:      3,
			Authors: entity.NewCollection(authors...),
		},
	}

	shop1 := &Shop{
		ID:      1,
		Profile: entity.NewCell(&ShopProfile{ID: 1}),
		Books:   entity.NewCollection(books...)}
	shop2 := &Shop{
		ID:      2,
		Profile: entity.NewCell(&ShopProfile{ID: 2}),
		Books: entity.NewCollection(&Book{
			ID: 2,
		}),
	}

	graph := createNewGraph()
	_ = graph.ProcessEntity(entity.NewBox(shop1, &meta))
	_ = graph.ProcessEntity(entity.NewBox(shop2, &meta))

	testStorage := &pusherStub{}
	executor := NewExecutor(testStorage, func(act CompositeAction) {})
	err := executor.Exec(graph)
	assert.NoError(t, err)

	testStorage.assertQueryAfter(t, queryStub{tableName: "shop", values: map[string]interface{}{"id": 1, "profile_id": 1}},
		queryStub{tableName: "shopprofile", values: map[string]interface{}{"id": 1}})

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

	testStorage.assertQueryAfter(t, queryStub{tableName: "shop", values: map[string]interface{}{"id": 2, "profile_id": 2}}, queryStub{tableName: "shopprofile", values: map[string]interface{}{"id": 2}})
	testStorage.assertQueryAfter(t, queryStub{tableName: "book", values: map[string]interface{}{"id": 2, "shop_id": 2}},
		queryStub{tableName: "shop", values: map[string]interface{}{"id": 2, "profile_id": 2}})
}

type Order struct {
	ID       int                `d3:"pk:auto"`
	Items    *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/OrderItem,join_on:order_id>,type:lazy"`
	BestItem *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/OrderItem,join_on:best_item_id>,type:lazy"`
}

func (o *Order) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*Order).ID, nil
				case "Items":
					return s.(*Order).Items, nil
				case "BestItem":
					return s.(*Order).BestItem, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*Order)
				e2T := e2.(*Order)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Items":
					return e1T.Items == e2T.Items
				case "BestItem":
					return e1T.BestItem == e2T.BestItem
				default:
					return false
				}
			},
		},
	}
}

type OrderItem struct {
	ID int `d3:"pk:auto"`
}

func (o *OrderItem) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*OrderItem).ID, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*OrderItem)
				e2T := e2.(*OrderItem)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				default:
					return false
				}
			},
		},
	}
}

func TestExecuteWithCircularReference(t *testing.T) {
	_ = metaRegistry.Add(
		(*Order)(nil),
		(*OrderItem)(nil),
	)

	bestItem := &OrderItem{ID: 1}
	orderItems := []interface{}{
		bestItem,
		&OrderItem{ID: 2},
	}

	order := &Order{
		ID:       1,
		Items:    entity.NewCollection(orderItems...),
		BestItem: entity.NewCell(bestItem),
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
