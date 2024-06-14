package structmap

import (
	"reflect"
	"strings"
)

// generates map[string]interface for the given struct
// for example:
//
//	test := &Test{
//		Id: 2,
//		A:  3,
//		B:  "4",
//		N:  &Nested{C: 5},
//	}
//	fields := []string{"A", "B", "C"}
//	expected := map[string]interface{}{
//		"A":   3,
//		"B":   "4",
//		"N.C": 5,  // or "C" : 5 if useDot == false
//	}
//
// it's useful for Update method
// if usePoint is set then name in format "A.B"
// GenerateFieldsMapFor generates a map of field names to their values for a given struct, including nested fields.
func GenerateFieldsMapFor(v any, fieldNames []string, useDot ...bool) map[string]interface{} {
	values := make(map[string]interface{})

	var processFields func(reflect.Value, string, []string)
	processFields = func(val reflect.Value, prefix string, fieldNames []string) {
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return
		}

		for fieldNum, field := range fieldNames {
			// Handle nested fields
			parts := strings.Split(field, ".")
			if len(parts) > 1 {
				// Process the prefix part of the field name
				fieldName := parts[0]
				f := val.FieldByName(fieldName)
				if f.IsValid() {
					// Process the rest of the field name recursively
					newPrefix := prefix
					if newPrefix != "" && len(useDot) > 0 && useDot[0] == true {
						newPrefix = newPrefix + "." + fieldName
					} else {
						newPrefix = fieldName
					}
					fieldNames[fieldNum] = parts[1]
					processFields(f, newPrefix, fieldNames)
				}
			} else {
				f := val.FieldByName(field)
				if !f.IsValid() {
					if val.Kind() == reflect.Struct {
						for i := 0; i < val.NumField(); i++ {
							subField := val.Field(i)
							subFieldName := val.Type().Field(i).Name
							processFields(subField, subFieldName, fieldNames)
						}
					}
					continue
				}
				fieldName := field
				if prefix != "" && len(useDot) > 0 && useDot[0] == true {
					fieldName = prefix + "." + field
				}
				if f.Kind() == reflect.Struct || f.Kind() == reflect.Ptr {
					processFields(f, fieldName, fieldNames)
				} else {
					values[fieldName] = f.Interface()
				}
			}
		}
	}

	processFields(reflect.ValueOf(v), "", fieldNames)
	return values
}
