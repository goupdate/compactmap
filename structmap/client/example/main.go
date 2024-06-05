package main

import (
	"log"
	"time"

	"github.com/goupdate/compactmap/structmap"
	"github.com/goupdate/compactmap/structmap/client"
	"github.com/stretchr/testify/assert"
)

type some1 struct {
	Id   int64  `json:"id"`
	Aba  string `json:"aba"`
	Haba int    `json:"haba"`
}

func main() {
	client.Timeout = 15 * time.Second

	client := client.New[some1]("http://localhost:80")

	// Clear the server storage
	err := client.Clear()
	if err != nil {
		log.Fatalf("Failed to clear storage: %v", err)
	}
	log.Println("Cleared storage")

	// Test Add
	item := &some1{Aba: "test", Haba: 12}
	id, err := client.Add(item)
	if err != nil {
		log.Fatalf("Failed to add item: %v", err)
	}
	log.Printf("Added item with ID %d\n", id)

	// Test Get
	retrievedItem, err := client.Get(id)
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}
	assert.Equal(nil, item.Aba, retrievedItem.Aba, "Field Aba should match")
	assert.Equal(nil, item.Haba, retrievedItem.Haba, "Field Haba should match")
	log.Printf("Retrieved item: %+v\n", retrievedItem)

	// Test Delete
	err = client.Delete(id)
	if err != nil {
		log.Fatalf("Failed to delete item: %v", err)
	}
	log.Println("Deleted item")

	val, err := client.Get(id)
	assert.Nil(nil, err, "Expected no err")
	assert.Nil(nil, val, "Expected not found item eq nil value")

	// Test Update
	item1 := &some1{Aba: "update_test", Haba: 50}
	id1, err := client.Add(item1)
	if err != nil {
		log.Fatalf("Failed to add item: %v", err)
	}

	conditions := []structmap.FindCondition{
		{Field: "Aba", Value: "update_test", Op: "equal"},
	}
	fields := map[string]interface{}{
		"Haba": 60,
	}
	updatedCount, err := client.Update("AND", conditions, fields)
	if err != nil {
		log.Fatalf("Failed to update item: %v", err)
	}
	assert.Equal(nil, 1, updatedCount, "Expected one item to be updated")
	log.Printf("Updated %d items\n", updatedCount)

	// --- updatecount
	item = &some1{Aba: "test", Haba: 12}
	item2 := &some1{Aba: "test2", Haba: 22}
	id, err = client.Add(item)
	if err != nil {
		log.Fatalf("Failed to add item: %v", err)
	}
	id, err = client.Add(item2)
	if err != nil {
		log.Fatalf("Failed to add item: %v", err)
	}

	conditions = []structmap.FindCondition{
		{Field: "Aba", Value: "test", Op: "contains"},
	}
	fields = map[string]interface{}{
		"Haba": 90,
	}
	updated, err := client.UpdateCount("AND", conditions, fields, 2)
	if err != nil {
		log.Fatalf("Failed to update item: %v", err)
	}
	if len(updated) != 2 {
		log.Fatalf("expected 2 ids: %+v\n", updated)
	}
	log.Printf("Updated %+v items\n", updated)

	// ----

	updatedItem, err := client.Get(id1)
	if err != nil {
		log.Fatalf("Failed to get updated item: %v", err)
	}
	assert.Equal(nil, 90, updatedItem.Haba, "Expected Haba to be updated")
	log.Printf("Updated item: %+v\n", updatedItem)

	// Test SetField
	err = client.SetField(id1, "Aba", "setfield_test")
	if err != nil {
		log.Fatalf("Failed to set field: %v", err)
	}

	updatedItem, err = client.Get(id1)
	if err != nil {
		log.Fatalf("Failed to get updated item: %v", err)
	}
	assert.Equal(nil, "setfield_test", updatedItem.Aba, "Expected Aba to be updated")
	log.Printf("Updated item after SetField: %+v\n", updatedItem)

	// Test SetFields
	fields = map[string]interface{}{
		"Aba":  "setfields_test",
		"Haba": 70,
	}
	err = client.SetFields(id1, fields)
	if err != nil {
		log.Fatalf("Failed to set fields: %v", err)
	}

	updatedItem, err = client.Get(id1)
	if err != nil {
		log.Fatalf("Failed to get updated item: %v", err)
	}
	assert.Equal(nil, "setfields_test", updatedItem.Aba, "Expected Aba to be updated")
	assert.Equal(nil, 70, updatedItem.Haba, "Expected Haba to be updated")
	log.Printf("Updated item after SetFields: %+v\n", updatedItem)

	// Test Find
	item2 = &some1{Aba: "find_test1", Haba: 80}
	item3 := &some1{Aba: "find_test2", Haba: 90}
	client.Add(item2)
	client.Add(item3)

	conditions = []structmap.FindCondition{
		{Field: "Aba", Value: "find_test1", Op: "equal"},
	}
	results, err := client.Find("AND", conditions)
	if err != nil {
		log.Fatalf("Failed to find items: %v", err)
	}
	assert.Equal(nil, 1, len(results), "Expected one item to be found")
	assert.Equal(nil, "find_test1", results[0].Aba, "Expected Aba to match")
	log.Printf("Found items: %+v\n", results)

	conditions = []structmap.FindCondition{
		{Field: "aba", Value: "find_test1", Op: "equal"},
	}
	results, err = client.Find("AND", conditions)
	if err != nil {
		log.Fatalf("Failed to find case-insensitive items: %v", err)
	}
	assert.Equal(nil, 1, len(results), "Expected one item to be found")
	assert.Equal(nil, "find_test1", results[0].Aba, "Expected Aba to match")
	log.Printf("Found case-insensitive items: %+v\n", results)

	// Test Iterate
	results, err = client.All()
	if err != nil {
		log.Fatalf("Failed to iterate items: %v", err)
	}
	assert.GreaterOrEqual(nil, len(results), 3, "Expected at least three items")
	log.Printf("Iterated items: %+v\n", results)
}
