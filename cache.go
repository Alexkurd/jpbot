package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Cache struct {
	Member []ChatMember
	// Add your actual data structure here
	// Ex: map[string]interface{}
	LastChanged int64 `json:"last_changed"`
	// ... other data fields
}

type ChatMember struct {
	Id            int64 `json:"id"`
	WelcomeShowed bool  `json:"welcome"`
	Rank          int   `json:"rank"`
	MessageCount  int   `json:"count"`
}

var (
	cache     Cache
	dataMutex sync.RWMutex
)

func loadCache() error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	file, err := os.ReadFile("cache.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No cache file.")
			// Create empty data if file doesn't exist
			cache = Cache{}
			return nil
		}
		return fmt.Errorf("error reading cache.json: %v", err)
	}

	err = json.Unmarshal(file, &cache)
	if err != nil {
		return fmt.Errorf("error unmarshalling cache.json: %v", err)
	}
	return nil
}

func saveCache() error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	file, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}

	err = os.WriteFile("cache.json", file, 0644)
	if err != nil {
		return fmt.Errorf("error saving data to cache.json: %v", err)
	}
	return nil
}

func importData() (bool, error) {
	err := loadCache()
	if err != nil {
		fmt.Println("Error loading data:", err)
	}
	// Replace with your actual data import logic
	// Return true if imported successfully, false otherwise
	// If successful, update data.LastChanged before returning
	fmt.Println("Placeholder for data import logic")

	fmt.Println("Last changed:", cache.LastChanged)
	cache.LastChanged = time.Now().Unix()
	return true, nil

	//return false, nil
}

func syncData() {
	for {
		// imported, err := importData()
		// if err != nil {
		// 	fmt.Println("Error importing data:", err)
		// } else if imported {
		// 	fmt.Println("Data imported successfully")
		// }

		// Update last_changed regardless of import status
		cache.LastChanged = time.Now().Unix()

		if err := saveCache(); err != nil {
			fmt.Println("Error saving data:", err)
		}
		time.Sleep(1 * time.Minute)
	}
}

func getMember(memberId int64) (member *ChatMember) {
	for _, member := range cache.Member {
		if member.Id == memberId {
			current := member
			return &current
		}
	}
	return nil
}
