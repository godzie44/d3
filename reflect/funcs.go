package reflect

import (
	"errors"
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
