// Code generated by d3. DO NOT EDIT.

package persist

import "fmt"
import "d3/orm/entity"

func (s *ShopCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl:       (*ShopCirc)(nil),
		TableName: "",
		Tools: entity.InternalTools{
			FieldExtractor: s.__d3_createFieldExtractor(),
		},
	}
}

func (s *ShopCirc) __d3_createFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*ShopCirc)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}

		switch name {

		case "Id":
			return sTyped.Id, nil

		case "Name":
			return sTyped.Name, nil

		case "Profile":
			return sTyped.Profile, nil

		case "FriendShop":
			return sTyped.FriendShop, nil

		case "TopSeller":
			return sTyped.TopSeller, nil

		case "Sellers":
			return sTyped.Sellers, nil

		case "KnownSellers":
			return sTyped.KnownSellers, nil

		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}

func (s *ShopProfileCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl:       (*ShopProfileCirc)(nil),
		TableName: "",
		Tools: entity.InternalTools{
			FieldExtractor: s.__d3_createFieldExtractor(),
		},
	}
}

func (s *ShopProfileCirc) __d3_createFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*ShopProfileCirc)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}

		switch name {

		case "Id":
			return sTyped.Id, nil

		case "Shop":
			return sTyped.Shop, nil

		case "Description":
			return sTyped.Description, nil

		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}

func (s *SellerCirc) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl:       (*SellerCirc)(nil),
		TableName: "",
		Tools: entity.InternalTools{
			FieldExtractor: s.__d3_createFieldExtractor(),
		},
	}
}

func (s *SellerCirc) __d3_createFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*SellerCirc)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}

		switch name {

		case "Id":
			return sTyped.Id, nil

		case "Name":
			return sTyped.Name, nil

		case "CurrentShop":
			return sTyped.CurrentShop, nil

		case "KnownShops":
			return sTyped.KnownShops, nil

		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}
