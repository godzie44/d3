package persistence

type Scanner interface {
	Scan(...interface{}) error
}

type (
	Pusher interface {
		Insert(table string, cols []string, values []interface{}, onConflict OnConflict) error
		InsertWithReturn(table string, cols []string, values []interface{}, returnCols []string, withReturned func(scanner Scanner) error) error
		Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error
		Remove(table string, identityCond map[string]interface{}) error
	}

	OnConflict int
)

const (
	_ OnConflict = iota
	Undefined
	DoNothing
)
