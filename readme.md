# CompactMap

CompactMap is a Go library that provides a memory-efficient alternative to the standard map with a comparable performance. It organizes entries into multiple buffers to optimize memory usage, making it suitable for applications where memory efficiency is critical.

## Features

- Memory-efficient storage for large datasets.
- Comparable performance to the standard map.
- Supports serialization and deserialization.
- Thread-safe operations with `sync.Mutex`.

## Installation

To install the CompactMap library, use `go get`:

```sh
go get github.com/goupdate/compactmap
```

## Usage

### Creating a CompactMap

To create a new CompactMap, use the `NewCompactMap` function:

```go
package main

import (
	"fmt"
	"github.com/goupdate/compactmap"
)

func main() {
	cm := compactmap.NewCompactMap[int, int]()
	cm.Add(1, 100)
	value, exists := cm.Get(1)
	if exists {
		fmt.Println("Value:", value)
	}
}
```

### Adding Entries

To add entries to the CompactMap, use the `Add` method:

```go
cm.Add(1, 100)
cm.Add(2, 200)
```

### Getting Entries

To retrieve entries from the CompactMap, use the `Get` method:

```go
value, exists := cm.Get(1)
if exists {
    fmt.Println("Value:", value)
}
```

### Deleting Entries

To delete entries from the CompactMap, use the `Delete` method:

```go
cm.Delete(1)
```

### Iterating Over Entries

To iterate over entries in the CompactMap, use the `Iterate` method:

```go
cm.Iterate(func(key, value int) bool {
    fmt.Printf("Key: %d, Value: %d\n", key, value)
    return true // return false to stop iteration
})
```

### Checking Existence of a Key

To check if a key exists in the CompactMap, use the `Exist` method:

```go
exists := cm.Exist(1)
fmt.Println("Exists:", exists)
```

### Counting Entries

To get the number of entries in the CompactMap, use the `Count` method:

```go
count := cm.Count()
fmt.Println("Count:", count)
```

### Saving and Loading

You can save the CompactMap to a file and load it later:

```go
// Save to file
err := cm.Save("compactmap.data")
if err != nil {
    fmt.Println("Error saving CompactMap:", err)
}

// Load from file
err = cm.Load("compactmap.data")
if err != nil {
    fmt.Println("Error loading CompactMap:", err)
}
```

## Performance

Here are the performance benchmarks for the CompactMap.
It's 2 times uses less memory of standart map with the same speed!

```
Memory used for standard map[int]int 10000000 elements = 170 MiB
CompactMem used for 10000000 elements = 97 MiB
stats: 10000 buffers, total len: 10000000
goos: windows
goarch: amd64
pkg: compactmap
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkStandardMap-12         15601668               176.9 ns/op
BenchmarkCompactMap-12          13738622               196.8 ns/op
PASS
ok      compactmap      8.217s
```

