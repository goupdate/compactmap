package compactmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndGet(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)
	cm.AddOrSet(3, 300)

	value, exists := cm.Get(1)
	assert.True(t, exists, "Value for key 1 should exist")
	assert.Equal(t, 100, value, "Value for key 1 should be 100")

	value, exists = cm.Get(2)
	assert.True(t, exists, "Value for key 2 should exist")
	assert.Equal(t, 200, value, "Value for key 2 should be 200")

	value, exists = cm.Get(3)
	assert.True(t, exists, "Value for key 3 should exist")
	assert.Equal(t, 300, value, "Value for key 3 should be 300")
}

func TestDelete(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)
	cm.Delete(1)

	_, exists := cm.Get(1)
	assert.False(t, exists, "Value for key 1 should not exist after deletion")

	value, exists := cm.Get(2)
	assert.True(t, exists, "Value for key 2 should exist")
	assert.Equal(t, 200, value, "Value for key 2 should be 200")
}

func TestIterate(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)
	cm.AddOrSet(3, 300)

	var result []int
	cm.Iterate(func(key, value int) bool {
		result = append(result, key)
		return true
	})

	assert.ElementsMatch(t, []int{1, 2, 3}, result, "Iterate should visit all keys")
}

func TestExist(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)

	exists := cm.Exist(1)
	assert.True(t, exists, "Key 1 should exist")

	exists = cm.Exist(3)
	assert.False(t, exists, "Key 3 should not exist")
}

func TestSortOrder(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(3, 300)
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)

	var result []int
	cm.Iterate(func(key, value int) bool {
		result = append(result, key)
		return true
	})

	assert.Equal(t, []int{1, 2, 3}, result, "Keys should be iterated in sorted order")
}

func TestSaveAndLoad(t *testing.T) {
	cm := NewCompactMap[int, int]()
	cm.AddOrSet(1, 100)
	cm.AddOrSet(2, 200)
	cm.Save("test_data.dat")

	cm2 := NewCompactMap[int, int]()
	cm2.Load("test_data.dat")

	value, exists := cm2.Get(1)
	assert.True(t, exists, "Value for key 1 should exist after load")
	assert.Equal(t, 100, value, "Value for key 1 should be 100 after load")

	value, exists = cm2.Get(2)
	assert.True(t, exists, "Value for key 2 should exist after load")
	assert.Equal(t, 200, value, "Value for key 2 should be 200 after load")
}

func TestErrorOnLoad(t *testing.T) {
	cm := NewCompactMap[int, int]()
	err := cm.Load("non_existent_file.dat")
	assert.Error(t, err, "Loading from a non-existent file should return an error")
}
