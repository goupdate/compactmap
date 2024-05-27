package compactmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	m := NewCompactMap[int32, int32](5)
	for i := int32(0); i < 10; i++ {
		m.Add(i, 100+i)
	}

	value, found := m.Get(5)
	assert.True(t, found)
	assert.True(t, value == 105)
	assert.True(t, m.Exist(5))

	m.Delete(5)
	_, found = m.Get(5)
	assert.True(t, !found)
	assert.True(t, m.Count() == 9)

	buf, _ := serialize(int64(1))
	d, _ := deserialize[int64](buf)
	assert.Equal(t, d, int64(1))
	assert.Equal(t, len(buf), 8)

	buf, _ = serialize(int32(2))
	d2, _ := deserialize[int32](buf)
	assert.Equal(t, d2, int32(2))
	assert.Equal(t, len(buf), 4)

	buf, _ = serialize(int16(3))
	d3, _ := deserialize[int16](buf)
	assert.Equal(t, d3, int16(3))
	assert.Equal(t, len(buf), 2)

	buf, _ = serialize(int8(4))
	d4, _ := deserialize[int8](buf)
	assert.Equal(t, d4, int8(4))
	assert.Equal(t, len(buf), 1)

	// Saving the map to a file
	err := m.Save("compactmap.dat")
	assert.True(t, err == nil)

	// Loading the map from a file
	m2 := NewCompactMap[int32, int32](5)
	err = m2.Load("compactmap.dat")
	if err != nil {
		t.Fatal("Error loading map:", err)
	}

	value, found = m2.Get(4)
	assert.True(t, found)
	assert.True(t, value == 104)
	assert.True(t, m2.Count() == 9)
}
