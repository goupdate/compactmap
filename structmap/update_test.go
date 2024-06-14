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

	result := GenerateFieldsMapFor(test, []string{"A", "B"}, false)

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}

func TestGenerateFieldsMapForNonPointer(t *testing.T) {
	test := TestStruct{Id: 2, A: 3, B: "4"}

	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
	}

	result := GenerateFieldsMapFor(test, []string{"A", "B"}, false)

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}

func TestGenerateFieldsMapForInvalidField(t *testing.T) {
	test := &TestStruct{Id: 2, A: 3, B: "4"}

	expected := map[string]interface{}{}

	result := GenerateFieldsMapFor(test, []string{"C"})

	assert.True(t, reflect.DeepEqual(expected, result), "Expected %v, but got %v", expected, result)
}

type Nested struct {
	C int
}

type Test struct {
	Id int64
	A  int
	B  string
	N  *Nested
}

func TestGenerateFieldsMapForNoDot(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{"A", "B", "N", "C"}
	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
		"C": 5, // equal "N.C"
	}

	result := GenerateFieldsMapFor(test, fields, false)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestGenerateFieldsMapForUseDot3(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{"A", "B", "N", "C"}
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

func TestGenerateFieldsMapForNoDot2(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{"A", "B", "C"}
	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
		"C": 5,
	}

	result := GenerateFieldsMapFor(test, fields)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestGenerateFieldsMapFor_EmptyFields(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{}
	expected := map[string]interface{}{}

	result := GenerateFieldsMapFor(test, fields)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestGenerateFieldsMapFor_NilPointer(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  nil,
	}
	fields := []string{"A", "B", "N"}
	expected := map[string]interface{}{
		"A": 3,
		"B": "4",
	}

	result := GenerateFieldsMapFor(test, fields)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestGenerateFieldsMapFor_NonStruct(t *testing.T) {
	nonStruct := 123
	fields := []string{"A", "B", "N"}
	expected := map[string]interface{}{}

	result := GenerateFieldsMapFor(nonStruct, fields)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
