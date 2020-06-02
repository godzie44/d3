package reflect

import (
	"errors"
	"reflect"
)

var (
	errElementNofFound = errors.New("element not found")
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
