package structmap

import (
	"testing"
)

func Test2FindBool(t *testing.T) {
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
		FindCondition{Field: "Field3", Value: true})
	if len(results) != 1 || results[0].Field1 != "value2" {
		t.Fatalf("expected to find 1 item with Field1 'value2', got %d items", len(results))
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
