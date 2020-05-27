package reflect

import (
	"errors"
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

func Copy(strctPtr interface{}) interface{} {
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
