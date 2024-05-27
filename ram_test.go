package compactmap

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRamUsage(t *testing.T) {
	var N = int32(1000 * 10000)

	keys := make([]int32, N)
	for i := int32(0); i < N; i++ {
		keys[i] = int32(rand.Intn(int(N)))
	}

	runtime.GC()
	var was runtime.MemStats
	runtime.ReadMemStats(&was)

	keys2 := make(map[int32]int32)
	for i := int32(0); i < N; i++ {
		keys2[i] = keys[i]
	}

	runtime.GC()
	var std runtime.MemStats
	runtime.ReadMemStats(&std)

	fmt.Printf("Memory used for standard map[int]int %dM elements = %v MiB\n", N, (std.Alloc-was.Alloc)/1024/1024)

	m := NewCompactMap[int32, int32]()
	for i := int32(0); i < N; i++ {
		m.Add(keys[i], i)
	}

	runtime.GC()
	var com runtime.MemStats
	runtime.ReadMemStats(&com)
	runtime.GC()

	fmt.Printf("CompactMem used for %dM elements = %v MiB\n", N, (com.Alloc-std.Alloc)/1024/1024)

	s := []*Entry[int32, int32]{}
	for i := int32(0); i < N; i++ {
		s = append(s, &Entry[int32, int32]{Key: i, Value: keys[i]})
	}

	runtime.GC()
	var sl runtime.MemStats
	runtime.ReadMemStats(&sl)

	fmt.Printf("Slice used for %dM elements = %v MiB\n", N, (sl.Alloc-com.Alloc)/1024/1024)
	assert.Equal(t, len(keys2), len(keys))

	fmt.Println(m.Count())
}
