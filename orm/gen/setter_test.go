package gen

import (
	"database/sql"
	"github.com/godzie44/d3/orm/query"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type setterTestStruct2 struct { //nolint
}

type setterTestStruct struct {
	int        int                //nolint
	intPtr     *int               //nolint
	string     string             //nolint
	setter2    *setterTestStruct2 //nolint
	closer     io.Closer          //nolint
	nullInt    sql.NullInt64      //nolint
	nullString sql.NullString     //nolint
	t          time.Time          //nolint
	tPtr       *time.Time         //nolint
	q          query.Query        //nolint
}

var expectedSetter = `func (s *setterTestStruct) __d3_makeFieldSetter() entity.FieldSetter {
	return func(s interface{}, name string, val interface{}) error {
		eTyped, ok := s.(*setterTestStruct)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		switch name { 
		case "int":
			eTyped.int = val.(int)
			return nil 
		case "intPtr":
			eTyped.intPtr = val.(*int)
			return nil 
		case "string":
			eTyped.string = val.(string)
			return nil 
		case "setter2":
			eTyped.setter2 = val.(*gen.setterTestStruct2)
			return nil 
		case "closer":
			eTyped.closer = val.(io.Closer)
			return nil 
		case "t":
			eTyped.t = val.(time.Time)
			return nil 
		case "tPtr":
			eTyped.tPtr = val.(*time.Time)
			return nil 
		case "q":
			eTyped.q = val.(query.Query)
			return nil 
		
		case "nullInt":
			if valuer, isValuer := val.(driver.Valuer); isValuer {
				v, err := valuer.Value()
				if err != nil {
					return eTyped.nullInt.Scan(nil)
				} 
				return eTyped.nullInt.Scan(v)
			}
			return eTyped.nullInt.Scan(val) 
		case "nullString":
			if valuer, isValuer := val.(driver.Valuer); isValuer {
				v, err := valuer.Value()
				if err != nil {
					return eTyped.nullString.Scan(nil)
				} 
				return eTyped.nullString.Scan(v)
			}
			return eTyped.nullString.Scan(val) 
		default:
			return fmt.Errorf("field %s not found", name)
		}
	}
}`

func TestSetterGeneration(t *testing.T) {
	buff := &strings.Builder{}
	gen := &setter{out: buff, imports: map[string]struct{}{}}

	gen.handle(reflect.TypeOf(setterTestStruct{}))

	assert.Equal(t, expectedSetter, strings.Trim(buff.String(), "\n"))

	imports := gen.preamble()
	sort.Strings(imports)
	assert.Equal(t, []string{"database/sql/driver", "github.com/godzie44/d3/orm/gen",
		"github.com/godzie44/d3/orm/query", "io", "time"}, imports)
}
