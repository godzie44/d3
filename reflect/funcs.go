package reflect

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/mohae/deepcopy"
	"reflect"
)

var (
	errElementNofFound = errors.New("element not found")
	ErrInvalidType     = errors.New("invalid type, must be struct or pointer to struct")
)

func GetFirstElementFromSlice(slice interface{}) (interface{}, error) {
	sliceVal := reflect.ValueOf(slice)

	if sliceVal.Len() < 1 {
		return nil, errElementNofFound
	}

	return sliceVal.Index(0).Interface(), nil
}

func CreateEmptyEntity(strctPtr interface{}) interface{} {
	entityType := reflect.TypeOf(strctPtr)
	return reflect.New(entityType.Elem()).Interface()
}

func CreateSliceOfStructPtrs(strctType reflect.Type, len int) interface{} {
	sliceType := reflect.SliceOf(strctType)
	sliceVal := reflect.MakeSlice(sliceType, len, len)

	return sliceVal.Interface()
}

func BreakUpSlice(slice interface{}) []interface{} {
	sliceVal := reflect.ValueOf(slice)
	var result = make([]interface{}, sliceVal.Len())

	for i := 0; i < sliceVal.Len(); i++ {
		result[i] = sliceVal.Index(i).Interface()
	}

	return result
}

func IntoStructType(t reflect.Type) (reflect.Type, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, ErrInvalidType
	}

	return t, nil
}

func FullName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return t.Elem().PkgPath() + "/" + t.Elem().Name()
	default:
		return t.PkgPath() + "/" + t.Name()
	}
}

func ExtractStructField(strctPtr interface{}, fieldName string) (interface{}, error) {
	elPtr := reflect.ValueOf(strctPtr)
	if elPtr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("entity value must be pointer")
	}

	if elPtr.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("entity pointer not link to struct")
	}

	field := elPtr.Elem().FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("invalid field: %s", fieldName)
	}

	return field.Interface(), nil
}

var scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func SetFields(strctPtr interface{}, fields map[string]interface{}) error {
	reflectVal := reflect.ValueOf(strctPtr).Elem()

	for name, val := range fields {
		f := reflectVal.FieldByName(name)

		if err := ValidateField(&f); err != nil {
			return err
		}

		SetField(&f, val)
	}

	return nil
}

func ValidateField(field *reflect.Value) error {
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("unreacheble field")
	}
	return nil
}

func SetField(field *reflect.Value, val interface{}) {
	if reflect.PtrTo(field.Type()).Implements(scannerType) {
		if valuer, isValuer := val.(driver.Valuer); isValuer {
			v, err := valuer.Value()
			if err != nil {
				_ = field.Addr().Interface().(sql.Scanner).Scan(nil)
			} else {
				_ = field.Addr().Interface().(sql.Scanner).Scan(v)
			}
		} else {
			_ = field.Addr().Interface().(sql.Scanner).Scan(val)
		}
	} else {
		field.Set(reflect.ValueOf(val))
	}
}

func Copy(strctPtr interface{}) interface{} {
	//val := reflect.ValueOf(strctPtr).Elem()
	//pt := reflect.PtrTo(val.Type())
	//pv := reflect.New(pt.Elem())
	//
	//pv.Elem().Set(val)
	//
	//return pv.Interface()

	return deepcopy.Copy(strctPtr)
}

func IsFieldEquals(strctPtr1, strctPtr2 interface{}, field string) bool {
	if strctPtr1 == nil || strctPtr2 == nil {
		return false
	}

	reflectVal1 := reflect.ValueOf(strctPtr1).Elem()
	reflectVal2 := reflect.ValueOf(strctPtr2).Elem()

	return reflect.DeepEqual(reflectVal1.FieldByName(field).Interface(), reflectVal2.FieldByName(field).Interface())
}
