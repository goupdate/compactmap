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

var cm = NewCompactMap[int, int]()

// Benchmark for CompactMap
func BenchmarkCompactMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cm.Add(rand.Intn(b.N), rand.Intn(b.N))
	}
}
