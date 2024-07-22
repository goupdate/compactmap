package structmap

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

type CustomString string

type ExampleStruct struct {
	Id     int64
	Field1 string
	Field2 int
	Field3 bool
	Field4 CustomString
	Field5 int64
	Field6 float64
}

func TestNew(t *testing.T) {
	storage, err := New[*ExampleStruct]("test_storage", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if storage == nil {
		t.Fatalf("expected non-nil storage")
	}
}

func TestAddAndGet(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	item, ok := storage.Get(id)
	if !ok {
		t.Fatalf("expected item, got nil")
	}
	if item.Field1 != "value1" || item.Field2 != 42 {
		t.Fatalf("expected %v, got %v", example, item)
	}
}

func TestSetField(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	updated := storage.SetField(id, "Field1", "newValue")
	if !updated {
		t.Fatalf("expected field to be updated")
	}

	item, ok := storage.Get(id)
	if !ok || item.Field1 != "newValue" {
		t.Fatalf("expected field1 to be 'newValue', got %v", item.Field1)
	}
}

func TestGetAll(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	items := storage.GetAll()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestFindEqual(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value1", Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}
}

func TestFindGreaterThan(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field2", Value: 42, Op: "gt"})
	if len(results) != 1 || results[0].Field2 != 43 {
		t.Fatalf("expected to find 1 item with Field2 greater than 42, got %d items", len(results))
	}
}

func TestFindLessThan(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field2", Value: 43, Op: "lt"})
	if len(results) != 1 || results[0].Field2 != 42 {
		t.Fatalf("expected to find 1 item with Field2 less than 43, got %d items", len(results))
	}
}

func TestFindLike(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 containing 'value', got %d items", len(results))
	}
}

func TestFindIn(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field2", Value: []int{40, 24, 42, 55}, Op: "in"})
	if len(results) != 1 {
		t.Fatalf("expected to find 1 items with Field1 containing 'value', got %d items", len(results))
	}

	results = storage.Find("", FindCondition{Field: "Field2", Value: []int{40, 42, 43, 44}, Op: "in"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 containing 'value', got %d items", len(results))
	}

	results = storage.Find("", FindCondition{Field: "Field2", Value: []int{40, 4, 5, 6, 7}, Op: "in"})
	if len(results) != 0 {
		t.Fatalf("expected to find 0 items with Field1 containing 'value', got %d items", len(results))
	}

	//string != int !
	results = storage.Find("", FindCondition{Field: "Field2", Value: []string{"40", "24", "42", "55"}, Op: "in"})
	if len(results) != 0 {
		t.Fatalf("expected to find 0 items with Field1 containing 'value', got %d items", len(results))
	}
}

func TestDelete(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	id1 := storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 containing 'value', got %d items", len(results))
	}

	storage.Delete(id1)
	_, ok := storage.Get(id1)
	if ok {
		t.Fatal("value should be removed")
	}
}

func TestClear(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) == 0 {
		t.Fatalf("expected to find items")
	}

	storage.Clear()

	items := storage.GetAll()
	if len(items) > 0 {
		t.Fatal("should be 0 items")
	}
}

func TestIterate(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "other", Field2: 44}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	var items []*ExampleStruct
	storage.Iterate(func(v *ExampleStruct) bool {
		items = append(items, v)
		return true
	})

	if len(items) != 3 {
		t.Fatalf("expected to iterate over 3 items, got %d items", len(items))
	}

	// Test stopping iteration early
	var partialItems []*ExampleStruct
	storage.Iterate(func(v *ExampleStruct) bool {
		partialItems = append(partialItems, v)
		return len(partialItems) < 2 // Stop after collecting 2 items
	})

	if len(partialItems) != 2 {
		t.Fatalf("expected to iterate over 2 items before stopping, got %d items", len(partialItems))
	}
}

func TestFindWithOrConditions(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "other", Field2: 42}
	example4 := &ExampleStruct{Field1: "value3", Field5: 45, Field2: 45}
	example5 := &ExampleStruct{Field1: "value3", Field5: 1721429365, Field2: 77}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)
	storage.Add(example4)
	storage.Add(example5)

	// Test OR condition
	results := storage.Find("OR", FindCondition{Field: "Field1", Value: "value1", Op: "equal"}, FindCondition{Field: "Field2", Value: 43, Op: "equal"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 'value1' or Field2 43, got %d items", len(results))
	}

	results = storage.Find("AND", FindCondition{Field: "Field5", Value: 41, Op: ">"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items, got %d items", len(results))
	}

	results = storage.Find("AND", FindCondition{Field: "Field2", Value: 44, Op: "<"})
	if len(results) != 3 {
		t.Fatalf("expected to find 3 items, got %d items", len(results))
	}

	results = storage.Find("AND", FindCondition{Field: "Field5", Value: 1721429361.0, Op: ">"})
	if len(results) != 1 {
		t.Fatalf("expected to find 1 items, got %d items", len(results))
	}
}

func TestFindWithAndConditions(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "value1", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	// Test AND condition
	results := storage.Find("AND", FindCondition{Field: "Field1", Value: "value1", Op: "equal"}, FindCondition{Field: "Field2", Value: 42, Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value1" || results[0].Field2 != 42 {
		t.Fatalf("expected to find 1 item with Field1 'value1' and Field2 42, got %d items", len(results))
	}
}

func TestFindIntAndFloat(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field6: 34}
	example2 := &ExampleStruct{Field1: "value2", Field5: 43}

	storage.Add(example1)
	storage.Add(example2)

	// Test AND condition
	results := storage.Find("AND", FindCondition{Field: "Field6", Value: int64(34), Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item, got %d items", len(results))
	}

	results = storage.Find("AND", FindCondition{Field: "Field5", Value: float64(43), Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value2" {
		t.Fatalf("expected to find 1 item, got %d items", len(results))
	}
}

func TestUpdate(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42, Id: 1}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43, Id: 2}
	example3 := &ExampleStruct{Field1: "value3", Field2: 44, Id: 3}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	// Update Field1 to "updated" for items where Field2 > 42
	updatedCount := storage.Update("AND", []FindCondition{
		{Field: "Field2", Value: 42, Op: "gt"},
	}, map[string]interface{}{
		"Field1": "updated",
	})

	if updatedCount != 2 {
		t.Fatalf("expected to update 2 items, got %d", updatedCount)
	}

	// Check if the updates were applied correctly
	results := storage.Find("OR", FindCondition{Field: "Id", Value: 2, Op: "equal"}, FindCondition{Field: "Id", Value: 3, Op: "equal"})
	for _, result := range results {
		if result.Field1 != "updated" {
			t.Fatalf("expected Field1 to be 'updated', got %s", result.Field1)
		}
	}
}

func TestUpdateCountRandom(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	for i := range 1000 {
		example1 := &ExampleStruct{Field1: "value" + fmt.Sprintf("%d", i), Field2: 42, Id: i}
		storage.Add(example1)
	}

	// Update Field1 to "updated" for items where Field2 > 42
	updatedCount := storage.UpdateCount("AND", []FindCondition{
		{Field: "Field2", Value: 42, Op: ">"},
	}, map[string]interface{}{
		"Field1": "updated",
	}, 1, true)

	if len(updatedCount) != 1 {
		t.Fatalf("expected to update 1 items, got %d", len(updatedCount))
	}

	updatedCountSecond := storage.UpdateCount("AND", []FindCondition{
		{Field: "Field2", Value: 42, Op: ">"},
	}, map[string]interface{}{
		"Field1": "updated2",
	}, 1, true)

	if len(updatedCountSecond) != 1 {
		t.Fatalf("expected to update 1 items, got %d", len(updatedCountSecond))
	}

	if updatedCount[0] == updatedCountSecond[0] {
		t.Fatalf("expected to update different elemtemts, but got same: %d", updatedCount[0])
	}
}
func TestSave(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	err := storage.Save()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storage2, _ := New[*ExampleStruct]("test_storage", false)
	example2, ok := storage2.Get(id)
	if !ok {
		t.Fatal("not found")
	}
	if *example != *example2 {
		t.Fatalf("fail on load after save: %+v vs %+v", example, example2)
	}

	// Clean up test files
	os.Remove("test_storage")
	os.Remove("test_storagei")
}

func TestMaxId(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	storage.Add(example)

	err := storage.Save()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storage2, _ := New[*ExampleStruct]("test_storage", false)
	id := storage2.GetMaxId()

	if id != 2 {
		t.Fatal("id should be = 2")
	}

	// Clean up test files
	os.Remove("test_storage")
	os.Remove("test_storagei")
}

func TestFindFieldByName(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}

	val := reflect.ValueOf(test).Elem()

	tests := []struct {
		name     string
		field    string
		expected interface{}
	}{
		{"Find nested field N.C", "N.C", 5},
		{"Find direct field A", "A", 3},
		{"Find non-existent field D", "D", nil},
		{"Find field in nested struct", "C", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := FindFieldByName(val, tt.field)
			if tt.expected == nil {
				if field.IsValid() {
					t.Errorf("Expected field %s to be invalid, but got valid field", tt.field)
				}
			} else {
				if !field.IsValid() {
					t.Errorf("Expected field %s to be valid, but got invalid field", tt.field)
				} else if !reflect.DeepEqual(field.Interface(), tt.expected) {
					t.Errorf("Expected %v, but got %v for field %s", tt.expected, field.Interface(), tt.field)
				}
			}
		})
	}
}

func TestGenerateFieldsMapFor2(t *testing.T) {
	test := &Test{
		Id: 2,
		A:  3,
		B:  "4",
		N:  &Nested{C: 5},
	}
	fields := []string{"A", "B", "C"}
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

func TestGenerateFieldsMapFor3(t *testing.T) {
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

func TestFindBool(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42, Field3: false}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43, Field3: true}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("",
		FindCondition{Field: "Field3", Value: false})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field1", Value: "value1"},
		FindCondition{Field: "Field1", Value: "e1", Op: "contains"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 1 || results[0].Field1 != "value2" {
		t.Fatalf("expected to find 1 item with Field1 'value2', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field1", Value: "value1"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field1", Value: []byte("value1")})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field2", Value: 13, Op: ">"},
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 1 || results[0].Field1 != "value2" {
		t.Fatalf("expected to find 1 item with Field1 'value2', got %d items", len(results))
	}

	results = storage.Find("AND",
		FindCondition{Field: "Field2", Value: 55, Op: ">"},
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 0 {
		t.Fatalf("expected to find 0 values, but got %d items : %+v", len(results), results[0])
	}

	results = storage.Find("",
		FindCondition{Field: "Field2", Value: 55, Op: ">"},
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 0 {
		t.Fatalf("expected to find 0 values, but got %d items : %+v", len(results), results[0])
	}

	results = storage.Find("OR",
		FindCondition{Field: "Field2", Value: 53, Op: ">"},
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 1 || results[0].Field1 != "value2" {
		t.Fatalf("expected to find 1 item with Field1 'value2', got %d items", len(results))
	}
}

func TestCompareValuesIn(t *testing.T) {
	tests := []struct {
		v1       interface{}
		v2       interface{}
		expected bool
	}{
		// Test for int
		{3, []int{1, 2, 3, 4, 5}, true},
		{6, []int{1, 2, 3, 4, 5}, false},

		// Test for string
		{"a", []string{"a", "b", "c"}, true},
		{"z", []string{"a", "b", "c"}, false},

		// Test for bool
		{true, []bool{true, false}, true},
		{false, []bool{true}, false},

		// Test for float
		{3.14, []float64{1.23, 3.14, 4.56}, true},
		{7.89, []float64{1.23, 3.14, 4.56}, false},
	}

	for _, tt := range tests {
		result := compareValues(tt.v1, tt.v2, "in")
		if result != tt.expected {
			t.Errorf("compareValues(%v, %v, \"in\") = %v; expected %v", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestFindEmptyString(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42, Field3: false}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43, Field3: true}
	example3 := &ExampleStruct{Field1: "", Field2: 55, Field3: true}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	results := storage.Find("",
		FindCondition{Field: "Field1", Value: "value1"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}

	results = storage.Find("",
		FindCondition{Field: "Field1"})
	if len(results) != 1 || results[0].Field2 != 55 {
		t.Fatalf("expected to find 1 item with Field2 '55', got %d items", len(results))
	}

}

func TestConvertToUnderlyingType(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"[]byte type", []byte("test"), "test"},
		{"String type", "test", "test"},
		{"Int type", 42, 42},
		{"Float type", 3.14, 3.14},
		{"Bool type", true, true},
		{"Slice type", []int{1, 2, 3}, []int{1, 2, 3}},
		{"CustomString type", CustomString("custom"), "custom"},
		{"Pointer to String", func() interface{} { s := "pointer"; return &s }(), "pointer"},
		{"Pointer to Int", func() interface{} { i := 42; return &i }(), 42},
		{"Pointer to CustomString", func() interface{} {
			cs := CustomString("custom pointer")
			return &cs
		}(),
			"custom pointer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := convertToUnderlyingType(reflect.ValueOf(tt.input))
			if !reflect.DeepEqual(val.Interface(), tt.expected) {
				t.Errorf("Expected %v (%v), but got %v (%v)", tt.expected, reflect.TypeOf(tt.expected).Kind().String(), val.Interface(), reflect.TypeOf(val.Interface()))
			}
		})
	}
}
