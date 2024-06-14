package structmap

import (
	"reflect"
	"testing"
)

func TestGenerateFieldsMapForUseDot(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{"A", "B", "C"} // "N" is incorrect
	expected := map[string]interface{}{
		"A":   3,
		"B":   "4",
		"N.C": 5,
	}

	result := GenerateFieldsMapFor(test, fields, true)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
