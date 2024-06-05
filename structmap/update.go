package structmap

import (
	"reflect"
)

// generates map[string]interface for the given struct
// for example:
//
//	type Test struct {
//		Id int64
//		A  int
//		B  string
//	}
//
// var test = &Test{Id:2, A:3, B:"4"}
// GenerateFieldsMapFor(test, []string{"A","B"} returns:
// map[] : {"A":3, "B":"4"}
// it's useful for Update method
func GenerateFieldsMapFor(v any, fieldNames []string) map[string]interface{} {
	// Check if v is a pointer and get the underlying element
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Initialize the map to store field values
	values := make(map[string]interface{})

	// Iterate over the field names and add their values to the map
	for _, field := range fieldNames {
		f := val.FieldByName(field)
		if f.IsValid() {
			values[field] = f.Interface()
		}
	}

	return values
}
