package persist

import (
	"d3/orm/entity"
	"database/sql"
)

//d3:entity
type ShopCirc struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string

	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/ShopProfileCirc,join_on:profile_id>,type:lazy"`

	FriendShop entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/ShopCirc,join_on:friend_id>,type:lazy"`

	TopSeller    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/SellerCirc,join_on:top_seller_id>,type:lazy"`
	Sellers      entity.Collection    `d3:"one_to_many:<target_entity:d3/tests/integration/persist/SellerCirc,join_on:shop_id>,type:lazy"`
	KnownSellers entity.Collection    `d3:"many_to_many:<target_entity:d3/tests/integration/persist/SellerCirc,join_on:shop_id,reference_on:seller_id,join_table:known_shop_seller_c>,type:lazy"`
}

//d3:entity
type ShopProfileCirc struct {
	Id          sql.NullInt32        `d3:"pk:auto"`
	Shop        entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/ShopCirc,join_on:shop_id>,type:lazy"`
	Description string
}

//d3:entity
type SellerCirc struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string

	CurrentShop entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/ShopCirc,join_on:shop_id>,type:lazy"`
	KnownShops  entity.Collection    `d3:"many_to_many:<target_entity:d3/tests/integration/persist/ShopCirc,join_on:seller_id,reference_on:shop_id,join_table:known_shop_seller_c>,type:lazy"`
}