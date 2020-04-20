package persistence

type Scanner interface {
	Scan(...interface{}) error
}

type Storage interface {
	Insert(table string, cols, pkCols []string, values []interface{}, propagatePk bool, propagationFn func(scanner Scanner) error) error
	Update(table string, cols []string, values []interface{}, identityCond map[string]interface{}) error
	Remove(table string, identityCond map[string]interface{}) error
}
