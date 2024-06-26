package compactmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	m := NewCompactMap[int32, int32]()
	for i := int32(0); i < 10; i++ {
		m.AddOrSet(i, 100+i)
	}

	value, found := m.Get(5)
	assert.True(t, found)
	assert.True(t, value == 105)
	assert.True(t, m.Exist(5))

	m.Delete(5)
	_, found = m.Get(5)
	assert.True(t, !found)
	assert.True(t, m.Count() == 9)

	buf, _ := Serialize(int64(1))
	d, _ := Deserialize[int64](buf)
	assert.Equal(t, d, int64(1))

	buf, _ = Serialize(int32(2))
	d2, _ := Deserialize[int32](buf)
	assert.Equal(t, d2, int32(2))

	buf, _ = Serialize(int16(3))
	d3, _ := Deserialize[int16](buf)
	assert.Equal(t, d3, int16(3))

	buf, _ = Serialize(int8(4))
	d4, _ := Deserialize[int8](buf)
	assert.Equal(t, d4, int8(4))

	// Saving the map to a file
	err := m.Save("dat")
	assert.True(t, err == nil)

	// Loading the map from a file
	m2 := NewCompactMap[int32, int32]()
	err = m2.Init("dat")
	if err != nil {
		t.Fatal("Error loading map:", err)
	}

	value, found = m2.Get(4)
	assert.True(t, found)
	assert.True(t, value == 104)
	assert.True(t, m2.Count() == 9)
}
