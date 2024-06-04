package main

import (
	"log"

	"github.com/goupdate/compactmap/structmap"

	"github.com/goupdate/compactmap/structmap/client"
)

type some1 struct {
	Id   int64  `json:"id"`
	Aba  string `json:"aba"`
	Haba int    `json:"haba"`
}

func main() {
	client := client.New[some1]("http://localhost:80")

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
	log.Printf("Retrieved item: %+v\n", retrievedItem)

	// Test Delete
	err = client.Delete(id)
	if err != nil {
		log.Fatalf("Failed to delete item: %v", err)
	}
	log.Println("Deleted item")

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
	log.Printf("Updated %d items\n", updatedCount)

	updatedItem, err := client.Get(id1)
	if err != nil {
		log.Fatalf("Failed to get updated item: %v", err)
	}
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
	log.Printf("Updated item after SetFields: %+v\n", updatedItem)

	// Test Find
	item2 := &some1{Aba: "find_test1", Haba: 80}
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
	log.Printf("Found items: %+v\n", results)

	// Test Iterate
	results, err = client.Iterate()
	if err != nil {
		log.Fatalf("Failed to iterate items: %v", err)
	}
	log.Printf("Iterated items: %+v\n", results)
}
