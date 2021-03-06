package persistence

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type Shop struct {
	ID      int                `d3:"pk:manual"`
	Books   *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/Book,join_on:shop_id>,type:lazy"`
	Profile *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/ShopProfile,join_on:profile_id>,type:lazy"`
}

func (s *Shop) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*Shop).ID, nil
				case "Books":
					return s.(*Shop).Books, nil
				case "Profile":
					return s.(*Shop).Profile, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*Shop)
				e2T := e2.(*Shop)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Books":
					return e1T.Books == e2T.Books
				case "Profile":
					return e1T.Profile == e2T.Profile
				default:
					return false
				}
			},
		},
	}
}

type ShopProfile struct {
	ID int `d3:"pk:manual"`
}

func (s *ShopProfile) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*ShopProfile).ID, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*ShopProfile)
				e2T := e2.(*ShopProfile)
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

type Book struct {
	ID      int                `d3:"pk:manual"`
	Authors *entity.Collection `d3:"many_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/Author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
}

func (b *Book) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*Book).ID, nil
				case "Authors":
					return s.(*Book).Authors, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*Book)
				e2T := e2.(*Book)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Authors":
					return e1T.Authors == e2T.Authors
				default:
					return false
				}
			},
		},
	}
}

type Author struct {
	ID int `d3:"pk:manual"`
}

func (a *Author) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*Author).ID, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*Author)
				e2T := e2.(*Author)
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

func initMetaRegistry() *entity.MetaRegistry {
	metaRegistry := entity.NewMetaRegistry()
	_ = metaRegistry.Add(
		(*ShopProfile)(nil),
		(*Shop)(nil),
		(*Book)(nil),
		(*Author)(nil),
	)
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
	shop := &Shop{ID: 1, Profile: entity.NewCell(&ShopProfile{ID: 3})}

	graph := createNewGraph()

	err := graph.ProcessEntity(entity.NewBox(shop, &meta))
	assert.NoError(t, err)

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 1)
	shopAction := aggRoots[0].(*InsertAction)
	assertTableName(t, "shopprofile", shopAction)
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
	shop := &Shop{ID: 1, Books: entity.NewCollection(books...)}

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
	books := &Book{ID: 1, Authors: entity.NewCollection(authors...)}

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
			Authors: entity.NewCollection(authors...),
		},
	}

	shop1 := &Shop{ID: 1, Profile: entity.NewCell(&ShopProfile{ID: 1}),
		Books: entity.NewCollection(books...)}
	shop2 := &Shop{ID: 2, Profile: entity.NewCell(&ShopProfile{ID: 2}),
		Books: entity.NewCollection(&Book{
			ID: 2,
		}),
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
		case action.TableName == "shopprofile" && action.Values["id"] == 1:
			profileAction1 = action
		case action.TableName == "shopprofile" && action.Values["id"] == 2:
			profileAction2 = action
		case action.TableName == "author" && action.Values["id"] == 1:
			authorAction1 = action
		case action.TableName == "author" && action.Values["id"] == 2:
			authorAction2 = action
		}
	}

	assertTableName(t, "shopprofile", profileAction1)
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

	assertTableName(t, "shopprofile", profileAction2)
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
	book1 := &Book{ID: 1, Authors: entity.NewCollection(e4s...)}
	book2 := &Book{ID: 2, Authors: entity.NewCollection(commonAuthor)}

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
	entity31 := &Book{ID: 1, Authors: entity.NewCollection(e4s...)}
	entity32 := &Book{ID: 2, Authors: entity.NewCollection(commonAuthor)}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity31, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity32, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(entity32, &meta)))
	assert.NoError(t, graph.ProcessEntity(entity.NewBox(commonAuthor, &meta)))

	assert.Len(t, graph.filterRoots(), 4)
}

type User struct {
	ID           int                `d3:"pk:auto"`
	Avatar       *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/Photo,join_on:avatar_id>,type:lazy"`
	GoodPhotos   *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/Photo,join_on:user_good_id>,type:lazy"`
	PrettyPhotos *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/Photo,join_on:user_pretty_id>,type:lazy"`
}

func (u *User) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*User).ID, nil
				case "Avatar":
					return s.(*User).Avatar, nil
				case "GoodPhotos":
					return s.(*User).GoodPhotos, nil
				case "PrettyPhotos":
					return s.(*User).PrettyPhotos, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*User)
				e2T := e2.(*User)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Avatar":
					return e1T.Avatar == e2T.Avatar
				case "GoodPhotos":
					return e1T.GoodPhotos == e2T.GoodPhotos
				case "PrettyPhotos":
					return e1T.PrettyPhotos == e2T.PrettyPhotos
				default:
					return false
				}
			},
		},
	}
}

type Photo struct {
	ID int `d3:"pk:auto"`
}

func (p *Photo) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*Photo).ID, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*Photo)
				e2T := e2.(*Photo)
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

func TestGenerationTwoOneToManyOnOneEntity(t *testing.T) {
	_ = metaRegistry.Add(
		(*User)(nil),
		(*Photo)(nil),
	)
	meta, _ := metaRegistry.GetMeta((*User)(nil))

	goodAndPrettyPhoto := &Photo{ID: 1}

	user := &User{ID: 1,
		GoodPhotos:   entity.NewCollection(goodAndPrettyPhoto, &Photo{ID: 2}),
		PrettyPhotos: entity.NewCollection(goodAndPrettyPhoto, &Photo{ID: 3}),
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
	_ = metaRegistry.Add(
		(*User)(nil),
		(*Photo)(nil),
	)
	meta, _ := metaRegistry.GetMeta((*User)(nil))

	goodAndAvatarPhoto := &Photo{ID: 1}

	user := &User{ID: 1,
		GoodPhotos: entity.NewCollection(goodAndAvatarPhoto),
		Avatar:     entity.NewCell(goodAndAvatarPhoto),
	}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(user, &meta)))
	assert.Len(t, graph.filterRoots(), 1)

	assertNoCycle(t, graph)
}

type BookCirc struct {
	ID         int                `d3:"pk:auto"`
	Authors    *entity.Collection `d3:"many_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/Author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	MainAuthor *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/Author,join_on:m_author_id>,type:lazy"`
}

func (b *BookCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*BookCirc).ID, nil
				case "Authors":
					return s.(*BookCirc).Authors, nil
				case "MainAuthor":
					return s.(*BookCirc).MainAuthor, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*BookCirc)
				e2T := e2.(*BookCirc)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Authors":
					return e1T.Authors == e2T.Authors
				case "MainAuthor":
					return e1T.MainAuthor == e2T.MainAuthor
				default:
					return false
				}
			},
		},
	}
}

func TestNoCycleManyToManyToOne(t *testing.T) {
	_ = metaRegistry.Add(
		(*Author)(nil),
		(*BookCirc)(nil),
	)
	meta, _ := metaRegistry.GetMeta((*BookCirc)(nil))

	mainAuthor := &Author{ID: 1}
	book := &BookCirc{
		ID: 1,
		Authors: entity.NewCollection(
			mainAuthor, &Author{ID: 2},
		),
		MainAuthor: entity.NewCell(mainAuthor),
	}

	graph := createNewGraph()

	assert.NoError(t, graph.ProcessEntity(entity.NewBox(book, &meta)))

	aggRoots := graph.filterRoots()
	assert.Len(t, aggRoots, 2)

	assertNoCycle(t, graph)
}

type shopCirc struct {
	ID      sql.NullInt32      `d3:"pk:auto"`
	Profile *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/shopProfileCirc,join_on:profile_id>"`
	Sellers *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/persistence/sellerCirc,join_on:shop_id>"`
	Name    string
}

func (s *shopCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*shopCirc).ID, nil
				case "Profile":
					return s.(*shopCirc).Profile, nil
				case "Sellers":
					return s.(*shopCirc).Sellers, nil
				case "Name":
					return s.(*shopCirc).Name, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*shopCirc)
				e2T := e2.(*shopCirc)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Profile":
					return e1T.Profile == e2T.Profile
				case "Sellers":
					return e1T.Sellers == e2T.Sellers
				case "Name":
					return e1T.Name == e2T.Name
				default:
					return false
				}
			},
		},
	}
}

type shopProfileCirc struct {
	ID   sql.NullInt32 `d3:"pk:auto"`
	Shop *entity.Cell  `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/shopCirc,join_on:shop_id>"`
}

func (s *shopProfileCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*shopProfileCirc).ID, nil
				case "Shop":
					return s.(*shopProfileCirc).Shop, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*shopProfileCirc)
				e2T := e2.(*shopProfileCirc)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Shop":
					return e1T.Shop == e2T.Shop
				default:
					return false
				}
			},
		},
	}
}

type sellerCirc struct {
	ID   sql.NullInt32 `d3:"pk:auto"`
	Shop *entity.Cell  `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/persistence/shopCirc,join_on:shop_id>"`
}

func (s *sellerCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*sellerCirc).ID, nil
				case "Shop":
					return s.(*sellerCirc).Shop, nil
				default:
					return nil, nil
				}
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*sellerCirc)
				e2T := e2.(*sellerCirc)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Shop":
					return e1T.Shop == e2T.Shop
				default:
					return false
				}
			},
		},
	}
}

func TestNoCycleOneToOne(t *testing.T) {
	_ = metaRegistry.Add(
		(*shopCirc)(nil),
		(*shopProfileCirc)(nil),
		(*sellerCirc)(nil),
	)

	meta, _ := metaRegistry.GetMeta((*shopCirc)(nil))

	profile := &shopProfileCirc{}
	shop := &shopCirc{
		Profile: entity.NewCell(profile),
	}
	profile.Shop = entity.NewCell(shop)

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
	_ = metaRegistry.Add(
		(*shopCirc)(nil),
		(*shopProfileCirc)(nil),
		(*sellerCirc)(nil),
	)
	meta, _ := metaRegistry.GetMeta((*shopCirc)(nil))

	seller1 := &sellerCirc{}
	seller2 := &sellerCirc{}
	seller3 := &sellerCirc{}
	shop1 := &shopCirc{
		ID:      sql.NullInt32{Int32: 1, Valid: true},
		Sellers: entity.NewCollection(seller1),
	}
	shop2 := &shopCirc{
		ID:      sql.NullInt32{Int32: 2, Valid: true},
		Sellers: entity.NewCollection(seller2, seller3),
	}
	seller1.Shop = entity.NewCell(shop1)
	seller2.Shop = entity.NewCell(shop2)
	seller3.Shop = entity.NewCell(shop2)

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
		case action.TableName == "shopcirc" && action.Values["id"].(sql.NullInt32).Int32 == 1:
			shopAction1 = action
		case action.TableName == "shopcirc" && action.Values["id"].(sql.NullInt32).Int32 == 2:
			shopAction2 = action
		}
	}

	assertTableName(t, "shopcirc", shopAction1)
	assertChildrenAndSubs(t, 2, 0, shopAction1)

	assert.IsType(t, (*UpdateAction)(nil), shopAction1.children()[0])
	assertTableName(t, "sellercirc", shopAction1.children()[0])
	assertChildrenAndSubs(t, 0, 2, shopAction1.children()[0])

	assertTableName(t, "sellercirc", shopAction1.children()[1])
	assertChildrenAndSubs(t, 1, 1, shopAction1.children()[1])

	assertOwnership(t, shopAction1.children()[1], shopAction1.children()[0])

	assertTableName(t, "shopcirc", shopAction2)
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
