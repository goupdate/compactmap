package compactmap

import (
	"fmt"
	"runtime"
	"testing"
)

func TestRamUsage(t *testing.T) {
	var N = int32(10 * 1000 * 1000)

	runtime.GC()
	var was runtime.MemStats
	runtime.ReadMemStats(&was)

	keys2 := make(map[int32]int32)
	for i := int32(0); i < N; i++ {
		keys2[i] = i * 2
	}

	runtime.GC()
	var std runtime.MemStats
	runtime.ReadMemStats(&std)

	fmt.Printf("Memory used for standard map[int]int %d elements = %v MiB\n", N, (std.Alloc-was.Alloc)/1024/1024)

	m := NewCompactMap[int32, int32]()
	for i := int32(0); i < N; i++ {
		m.AddOrSet(i, i*2)
	}

	runtime.GC()
	var com runtime.MemStats
	runtime.ReadMemStats(&com)
	runtime.GC()

	fmt.Printf("CompactMem used for %d elements = %v MiB\n", N, (com.Alloc-std.Alloc)/1024/1024)

	if m.Count()+len(keys2) == 2 {
		t.Fatal()
	}

	fmt.Printf("stats: %s\n", m.Stats())
}
