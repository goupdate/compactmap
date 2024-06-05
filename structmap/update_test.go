package structmap

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Id int64
	A  int
	B  string
}

func TestGenerateFieldsMapFor(t *testing.T) {
	test := &TestStruct{Id: 2, A: 3, B: "4"}

	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
	}

	result := GenerateFieldsMapFor(test, []string{"A", "B"})

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}

func TestGenerateFieldsMapForNonPointer(t *testing.T) {
	test := TestStruct{Id: 2, A: 3, B: "4"}

	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
	}

	result := GenerateFieldsMapFor(test, []string{"A", "B"})

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}

func TestGenerateFieldsMapForInvalidField(t *testing.T) {
	test := &TestStruct{Id: 2, A: 3, B: "4"}

	expected := map[string]interface{}{}

	result := GenerateFieldsMapFor(test, []string{"C"})

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}
