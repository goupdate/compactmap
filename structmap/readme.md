# StructMap

StructMap is an extended version of `compactmap`, providing an in-memory database with additional functionality. It supports storing, retrieving, and searching for pointers to structs, with the ability to set fields, iterate through entries, and perform conditional searches using logical operators.

## Features

- Add, get, delete, and clear entries.
- Set struct fields dynamically.
- Iterate through all entries.
- Search entries with multiple conditions using logical operators (`AND`, `OR`).

## Installation

To install the package, use:

```sh
go get github.com/goupdate/compactmap/structmap
```

## Usage

Here's an example of how to use StructMap:

```go
package main

import (
	"fmt"
	"github.com/goupdate/structmap"
)

type ExampleStruct struct {
	Field1 string
	Field2 int
}

func main() {
	storage, err := structmap.New[*ExampleStruct]("example_storage", false)
	if err != nil {
		panic(err)
	}

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("AND", structmap.FindCondition{Field: "Field1", Value: "value1", Op: "equal"})
	for _, result := range results {
		fmt.Printf("Found: %+v\n", result)
	}
}
```

## API

### New

```go
func New[V any](storageFile string, failIfNotLoaded bool) (*StructMap[V], error)
```

Creates a new StructMap instance.

### Add

```go
func (p *StructMap[V]) Add(v V) int64
```

Adds a new entry to the StructMap.

### Get

```go
func (p *StructMap[V]) Get(id int64) (V, bool)
```

Gets an entry by its ID.

### Delete

```go
func (p *StructMap[V]) Delete(id int64)
```

Deletes an entry by its ID.

### Clear

```go
func (p *StructMap[V]) Clear()
```

Clears all entries from the StructMap.

### SetField

```go
func (p *StructMap[V]) SetField(id int64, field string, value interface{}) bool
```

Sets a field of an entry dynamically.

### GetAll

```go
func (p *StructMap[V]) GetAll() []V
```

Retrieves all entries from the StructMap.

### Find

```go
func (p *StructMap[V]) Find(condition string, where ...FindCondition) []V
```

Finds entries based on multiple conditions.

### Iterate

```go
func (p *StructMap[V]) Iterate(fn func(v V) bool)
```

Iterates through all entries.

### Save

```go
func (p *StructMap[V]) Save() error
```

Saves the current state of the StructMap.