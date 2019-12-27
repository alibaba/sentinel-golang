package datasource


import (
	"encoding/json"
	"reflect"
)

func JsonParseArray(rule interface{}) Parser {
	typ := reflect.TypeOf(rule)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return func(data []byte) (interface{}, error) {
		val := reflect.New(typ).Interface()
		err := json.Unmarshal(data, val)
		return val, err
	}
}

type Parser func([]byte) (interface{}, error)
func (fn Parser) Convert(data []byte) (interface{}, error) {
	return fn(data)
}

type Converter interface {
	Convert([]byte) (interface{}, error)
}

