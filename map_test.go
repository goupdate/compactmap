package compactmap

import (
	"math/rand"
	"testing"
)

// Benchmark for standard map
func BenchmarkStandardMap(b *testing.B) {
	m := make(map[int]int, b.N)
	for i := 0; i < b.N; i++ {
		m[rand.Intn(b.N)] = rand.Intn(b.N)
	}
}

// Benchmark for CompactMap
func BenchmarkCompactMap(b *testing.B) {
	cm := NewCompactMap[int, int]()
	for i := 0; i < b.N; i++ {
		cm.AddOrSet(rand.Intn(b.N), rand.Intn(b.N))
	}
}
