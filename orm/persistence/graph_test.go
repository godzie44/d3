package persistence

import (
	"d3/orm/entity"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type Shop struct {
	entity  struct{}             `d3:"table_name:shop"` //nolint:unused,structcheck
	ID      int                  `d3:"pk:manual"`
	Books   entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/persistence/Book,join_on:shop_id>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/ShopProfile,join_on:profile_id>,type:lazy"`
}

type ShopProfile struct {
	entity struct{} `d3:"table_name:profile"` //nolint:unused,structcheck
	ID     int      `d3:"pk:manual"`
}

type Book struct {
	entity  struct{}          `d3:"table_name:book"` //nolint:unused,structcheck
	ID      int               `d3:"pk:manual"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/orm/persistence/Author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
}

type Author struct {
	entity struct{} `d3:"table_name:author"` //nolint:unused,structcheck
	ID     int      `d3:"pk:manual"`
}

func initMetaRegistry() *entity.MetaRegistry {
	metaRegistry := entity.NewMetaRegistry()
	_ = metaRegistry.Add((*ShopProfile)(nil), (*Shop)(nil), (*Book)(nil), (*Author)(nil))
	return metaRegistry
}

var metaRegistry = initMetaRegistry()

func assertTableName(t *testing.T, tableName string, action CompositeAction) {
	switch a := action.(type) {
	case *InsertAction:
		assert.Equal(t, tableName, a.TableName)
	case *UpdateAction:
		assert.Equal(t, tableName, a.TableName)
	}
}

func assertChildrenAndSubs(t *testing.T, childCount, subCount int, action CompositeAction) {
	assert.Len(t, action.children(), childCount)
	assert.Equal(t, subCount, action.subscriptionsCount())
}

func assertOwnership(t *testing.T, parent, child CompositeAction) {
	assert.Contains(t, parent.children(), child)
}

func createNewGraph() *PersistGraph {
	return NewPersistGraph(func(b *entity.Box) (bool, error) {
		return false, nil
	}, func(box *entity.Box) interface{} {
		return nil
	})
}

func TestGenerationOneToOneEntity(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Shop)(nil))
	shop := &Shop{ID: 1, Profile: entity.NewWrapEntity(&ShopProfile{ID: 3})}

	graph := createNewGraph()

	err := graph.ProcessEntity(entity.NewBox(shop, &meta))
	assert.NoError(t, err)

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 1)
	shopAction := aggRoots[0].(*InsertAction)
	assertTableName(t, "profile", shopAction)
	assertChildrenAndSubs(t, 1, 0, shopAction)
	assertTableName(t, "shop", shopAction.children()[0])
	assertChildrenAndSubs(t, 0, 1, shopAction.children()[0])
}

func TestGenerationOneToManyEntity(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Shop)(nil))

	books := []interface{}{
		&Book{
			ID: 1,
		},
		&Book{
			ID: 3,
		},
	}
	shop := &Shop{ID: 1, Books: entity.NewCollection(books)}

	graph := createNewGraph()

	err := graph.ProcessEntity(entity.NewBox(shop, &meta))
	assert.NoError(t, err)

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 1)

	shopAction := aggRoots[0].(*InsertAction)
	assertTableName(t, "shop", shopAction)
	assertChildrenAndSubs(t, 2, 0, shopAction)

	for _, child := range shopAction.children() {
		assertTableName(t, "book", child)
		assertChildrenAndSubs(t, 0, 1, child)
	}
}

func TestGenerationManyToMany(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Book)(nil))

	authors := []interface{}{
		&Author{
			ID: 1,
		},
		&Author{
			ID: 2,
		},
	}
	books := &Book{ID: 1, Authors: entity.NewCollection(authors)}

	graph := createNewGraph()

	err := graph.ProcessEntity(entity.NewBox(books, &meta))
	assert.NoError(t, err)

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 3)

	//unmap unordered roots
	var bookAction, authorAction1, authorAction2 *InsertAction
	for _, root := range aggRoots {
		action := root.(*InsertAction)
		switch {
		case action.TableName == "book":
			bookAction = action
		case action.TableName == "author" && action.Values["id"] == 1:
			authorAction1 = action
		case action.TableName == "author" && action.Values["id"] == 2:
			authorAction2 = action
		}
	}

	assertTableName(t, "book", bookAction)
	assertChildrenAndSubs(t, 2, 0, bookAction)
	assertTableName(t, "author", authorAction1)
	assertChildrenAndSubs(t, 1, 0, authorAction1)
	assertTableName(t, "book_author", authorAction1.children()[0])
	assertChildrenAndSubs(t, 0, 2, authorAction1.children()[0])
	assertTableName(t, "author", authorAction2)
	assertChildrenAndSubs(t, 1, 0, authorAction2)
	assertTableName(t, "book_author", authorAction2.children()[0])
	assertChildrenAndSubs(t, 0, 2, authorAction2.children()[0])
}

func TestGenerationBigGraph(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Shop)(nil))

	authors := []interface{}{
		&Author{
			ID: 1,
		},
		&Author{
			ID: 2,
		},
	}

	books := []interface{}{
		&Book{
			ID: 1,
		},
		&Book{
			ID:      3,
			Authors: entity.NewCollection(authors),
		},
	}

	shop1 := &Shop{ID: 1, Profile: entity.NewWrapEntity(&ShopProfile{ID: 1}),
		Books: entity.NewCollection(books)}
	shop2 := &Shop{ID: 2, Profile: entity.NewWrapEntity(&ShopProfile{ID: 2}),
		Books: entity.NewCollection([]interface{}{&Book{
			ID: 2,
		}}),
	}

	graph := createNewGraph()

	err := graph.ProcessEntity(entity.NewBox(shop1, &meta))
	assert.NoError(t, err)

	err = graph.ProcessEntity(entity.NewBox(shop2, &meta))
	assert.NoError(t, err)

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 4)

	//unmap unordered roots
	var profileAction1, profileAction2, authorAction1, authorAction2 *InsertAction
	for _, root := range aggRoots {
		action := root.(*InsertAction)
		switch {
		case action.TableName == "profile" && action.Values["id"] == 1:
			profileAction1 = action
		case action.TableName == "profile" && action.Values["id"] == 2:
			profileAction2 = action
		case action.TableName == "author" && action.Values["id"] == 1:
			authorAction1 = action
		case action.TableName == "author" && action.Values["id"] == 2:
			authorAction2 = action
		}
	}

	assertTableName(t, "profile", profileAction1)
	assertChildrenAndSubs(t, 1, 0, profileAction1)

	shopAction1 := profileAction1.children()[0]
	assertTableName(t, "shop", shopAction1)
	assertChildrenAndSubs(t, 2, 1, shopAction1)

	shop1BookActions := getOrderedChildrenBy(shopAction1, "id")
	assertTableName(t, "book", shop1BookActions[0])
	assertChildrenAndSubs(t, 0, 1, shop1BookActions[0])
	assertTableName(t, "book", shop1BookActions[1])
	assertChildrenAndSubs(t, 2, 1, shop1BookActions[1])

	assertTableName(t, "author", authorAction1)
	assertChildrenAndSubs(t, 1, 0, authorAction1)

	assertTableName(t, "book_author", authorAction1.children()[0])
	assertChildrenAndSubs(t, 0, 2, authorAction1.children()[0])

	book2RelToAuthorActions := getOrderedChildrenBy(shop1BookActions[1], "author_id")

	assert.Equal(t, authorAction1.children()[0], book2RelToAuthorActions[0])

	assertTableName(t, "author", authorAction2)
	assertChildrenAndSubs(t, 1, 0, authorAction2)
	assert.Equal(t, authorAction2.children()[0], book2RelToAuthorActions[1])
	assertTableName(t, "book_author", authorAction2.children()[0])
	assertChildrenAndSubs(t, 0, 2, authorAction2.children()[0])

	assertTableName(t, "profile", profileAction2)
	assertChildrenAndSubs(t, 1, 0, profileAction2)
	assertTableName(t, "shop", profileAction2.children()[0])
	assertChildrenAndSubs(t, 1, 1, profileAction2.children()[0])
	assertTableName(t, "book", profileAction2.children()[0].children()[0])
	assertChildrenAndSubs(t, 0, 1, profileAction2.children()[0].children()[0])
}

func TestGenerationManyToManyWithCommonEntity(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Book)(nil))

	commonAuthor := &Author{
		ID: 1,
	}
	e4s := []interface{}{
		commonAuthor,
		&Author{
			ID: 2,
		},
	}
	book1 := &Book{ID: 1, Authors: entity.NewCollection(e4s)}
	book2 := &Book{ID: 2, Authors: entity.NewCollection([]interface{}{commonAuthor})}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(book1, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(book2, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 4)

	//unmap unordered roots
	var bookAction1, bookAction2, authorAction1, authorAction2 *InsertAction
	for _, root := range aggRoots {
		action := root.(*InsertAction)
		switch {
		case action.TableName == "book" && action.Values["id"] == 1:
			bookAction1 = action
		case action.TableName == "book" && action.Values["id"] == 2:
			bookAction2 = action
		case action.TableName == "author" && action.Values["id"] == 1:
			authorAction1 = action
		case action.TableName == "author" && action.Values["id"] == 2:
			authorAction2 = action
		}
	}

	assertTableName(t, "book", bookAction1)
	assertChildrenAndSubs(t, 2, 0, bookAction1)
	assertTableName(t, "author", authorAction1)
	assertChildrenAndSubs(t, 2, 0, authorAction1)

	book1RelationToAuthorActions := getOrderedChildrenBy(bookAction1, "author_id")

	assert.Equal(t, authorAction1.children()[0], book1RelationToAuthorActions[0])
	assertTableName(t, "book_author", book1RelationToAuthorActions[0])
	assertChildrenAndSubs(t, 0, 2, book1RelationToAuthorActions[0])
	assertTableName(t, "author", authorAction2)
	assertChildrenAndSubs(t, 1, 0, authorAction2)
	assertTableName(t, "book_author", book1RelationToAuthorActions[1])
	assertChildrenAndSubs(t, 0, 2, book1RelationToAuthorActions[1])

	assertTableName(t, "book", bookAction2)
	assertChildrenAndSubs(t, 1, 0, bookAction2)
	assert.Equal(t, authorAction1.children()[1], bookAction2.children()[0])
	assertTableName(t, "book_author", bookAction2.children()[0])
	assertChildrenAndSubs(t, 0, 2, bookAction2.children()[0])
}

func TestDoublePersistNotAffectGraph(t *testing.T) {
	meta, _ := metaRegistry.GetMeta((*Book)(nil))

	commonAuthor := &Author{
		ID: 1,
	}
	e4s := []interface{}{
		commonAuthor,
		&Author{
			ID: 3,
		},
	}
	entity31 := &Book{ID: 1, Authors: entity.NewCollection(e4s)}
	entity32 := &Book{ID: 2, Authors: entity.NewCollection([]interface{}{commonAuthor})}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity31, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity32, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity32, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(commonAuthor, &meta)))

	assert.Len(t, graph.filterRoots(), 4)
}

type User struct {
	entity       struct{}             `d3:"table_name:user"` //nolint:unused,structcheck
	Id           int                  `d3:"pk:auto"`
	Avatar       entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/Photo,join_on:avatar_id>,type:lazy"`
	GoodPhotos   entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/persistence/Photo,join_on:user_good_id>,type:lazy"`
	PrettyPhotos entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/persistence/Photo,join_on:user_pretty_id>,type:lazy"`
}

type Photo struct {
	entity struct{} `d3:"table_name:photo"` //nolint:unused,structcheck
	Id     int      `d3:"pk:auto"`
}

func TestGenerationTwoOneToManyOnOneEntity(t *testing.T) {
	_ = metaRegistry.Add((*User)(nil), (*Photo)(nil))
	meta, _ := metaRegistry.GetMeta((*User)(nil))

	goodAndPrettyPhoto := &Photo{Id: 1}

	user := &User{Id: 1,
		GoodPhotos:   entity.NewCollection([]interface{}{goodAndPrettyPhoto, &Photo{Id: 2}}),
		PrettyPhotos: entity.NewCollection([]interface{}{goodAndPrettyPhoto, &Photo{Id: 3}}),
	}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(user, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 1)

	assertNoCycle(t, graph)

	assert.Len(t, aggRoots[0].children(), 4)

	var insertIds []int
	for _, act := range aggRoots[0].children() {
		switch a := act.(type) {
		case *InsertAction:
			insertIds = append(insertIds, a.Values["id"].(int))
			assert.Equal(t, "photo", a.TableName)
		case *UpdateAction:
			assert.Equal(t, "photo", a.TableName)
		}
	}

	sort.Ints(insertIds)
	assert.Equal(t, []int{1, 2, 3}, insertIds)
}

func TestNoCycleOneToOneToMany(t *testing.T) {
	_ = metaRegistry.Add((*User)(nil), (*Photo)(nil))
	meta, _ := metaRegistry.GetMeta((*User)(nil))

	goodAndAvatarPhoto := &Photo{Id: 1}

	user := &User{Id: 1,
		GoodPhotos: entity.NewCollection([]interface{}{goodAndAvatarPhoto}),
		Avatar:     entity.NewWrapEntity(goodAndAvatarPhoto),
	}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(user, &meta)))
	assert.Len(t, graph.filterRoots(), 1)

	assertNoCycle(t, graph)
}

type BookCirc struct {
	entity     struct{}             `d3:"table_name:book2"` //nolint:unused,structcheck
	Id         int                  `d3:"pk:auto"`
	Authors    entity.Collection    `d3:"many_to_many:<target_entity:d3/orm/persistence/Author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	MainAuthor entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/Author,join_on:m_author_id>,type:lazy"`
}

func TestNoCycleManyToManyToOne(t *testing.T) {
	_ = metaRegistry.Add((*BookCirc)(nil), (*Author)(nil))
	meta, _ := metaRegistry.GetMeta((*BookCirc)(nil))

	mainAuthor := &Author{ID: 1}
	book := &BookCirc{
		Id: 1,
		Authors: entity.NewCollection([]interface{}{
			mainAuthor, &Author{ID: 2},
		}),
		MainAuthor: entity.NewWrapEntity(mainAuthor),
	}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(book, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 2)

	assertNoCycle(t, graph)
}

type shopCirc struct {
	entity  struct{}             `d3:"table_name:shop"` //nolint:unused,structcheck
	Id      sql.NullInt32        `d3:"pk:auto"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/shopProfileCirc,join_on:profile_id>"`
	Sellers entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/persistence/sellerCirc,join_on:shop_id>"`
	Name    string
}

type shopProfileCirc struct {
	entity struct{}             `d3:"table_name:profile"` //nolint:unused,structcheck
	Id     sql.NullInt32        `d3:"pk:auto"`
	Shop   entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/shopCirc,join_on:shop_id>"`
}

type sellerCirc struct {
	entity struct{}             `d3:"table_name:seller"` //nolint:unused,structcheck
	Id     sql.NullInt32        `d3:"pk:auto"`
	Shop   entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/persistence/shopCirc,join_on:shop_id>"`
}

func TestNoCycleOneToOne(t *testing.T) {
	_ = metaRegistry.Add((*shopCirc)(nil), (*shopProfileCirc)(nil), (*sellerCirc)(nil))
	meta, _ := metaRegistry.GetMeta((*shopCirc)(nil))

	profile := &shopProfileCirc{}
	shop := &shopCirc{
		Profile: entity.NewWrapEntity(profile),
	}
	profile.Shop = entity.NewWrapEntity(shop)

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(shop, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 1)

	assertNoCycle(t, graph)

}

func assertNoCycle(t *testing.T, graph *PersistGraph) {
	assert.False(t, hasCycle(graph), "cycle found")
}

func TestNoCycleOneToMany(t *testing.T) {
	_ = metaRegistry.Add((*shopCirc)(nil), (*shopProfileCirc)(nil), (*sellerCirc)(nil))
	meta, _ := metaRegistry.GetMeta((*shopCirc)(nil))

	seller1 := &sellerCirc{}
	seller2 := &sellerCirc{}
	seller3 := &sellerCirc{}
	shop1 := &shopCirc{
		Id:      sql.NullInt32{Int32: 1, Valid: true},
		Sellers: entity.NewCollection([]interface{}{seller1}),
	}
	shop2 := &shopCirc{
		Id:      sql.NullInt32{Int32: 2, Valid: true},
		Sellers: entity.NewCollection([]interface{}{seller2, seller3}),
	}
	seller1.Shop = entity.NewWrapEntity(shop1)
	seller2.Shop = entity.NewWrapEntity(shop2)
	seller3.Shop = entity.NewWrapEntity(shop2)

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(shop1, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(shop2, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 2)

	assertNoCycle(t, graph)

	var shopAction1, shopAction2 *InsertAction
	for _, root := range aggRoots {
		action := root.(*InsertAction)
		switch {
		case action.TableName == "shop" && action.Values["id"].(sql.NullInt32).Int32 == 1:
			shopAction1 = action
		case action.TableName == "shop" && action.Values["id"].(sql.NullInt32).Int32 == 2:
			shopAction2 = action
		}
	}

	assertTableName(t, "shop", shopAction1)
	assertChildrenAndSubs(t, 2, 0, shopAction1)

	assert.IsType(t, (*UpdateAction)(nil), shopAction1.children()[0])
	assertTableName(t, "seller", shopAction1.children()[0])
	assertChildrenAndSubs(t, 0, 2, shopAction1.children()[0])

	assertTableName(t, "seller", shopAction1.children()[1])
	assertChildrenAndSubs(t, 1, 1, shopAction1.children()[1])

	assertOwnership(t, shopAction1.children()[1], shopAction1.children()[0])

	assertTableName(t, "shop", shopAction2)
	assertChildrenAndSubs(t, 4, 0, shopAction2)
	assert.IsType(t, (*UpdateAction)(nil), shopAction2.children()[0])
	assert.IsType(t, (*UpdateAction)(nil), shopAction2.children()[2])

	assertOwnership(t, shopAction2.children()[1], shopAction2.children()[0])
	assertOwnership(t, shopAction2.children()[3], shopAction2.children()[2])
}

func getOrderedChildrenBy(act CompositeAction, sortField string) []CompositeAction {
	unwrapPromiseIfNeeded := func(v interface{}) int {
		if prom, ok := v.(*promise); ok {
			res, _ := prom.unwrap()
			return res.(int)
		}
		return v.(int)
	}

	children := act.children()
	sort.Slice(children, func(i, j int) bool {
		var idi, idj int
		switch act := children[i].(type) {
		case *InsertAction:
			idi = unwrapPromiseIfNeeded(act.Values[sortField])
		case *UpdateAction:
			idi = unwrapPromiseIfNeeded(act.Values[sortField])
		}

		switch act := children[j].(type) {
		case *InsertAction:
			idj = unwrapPromiseIfNeeded(act.Values[sortField])
		case *UpdateAction:
			idj = unwrapPromiseIfNeeded(act.Values[sortField])
		}

		return idi < idj
	})
	return children
}
