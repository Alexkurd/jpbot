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
	Member      []ChatMember
	DeleteList  []WelcomeMessage `json:"DeleteList,omitempty"`
	LastChanged int64            `json:"last_changed"`
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

func importCache() (bool, error) {
	err := loadCache()
	if err != nil {
		fmt.Println("Error loading data:", err)
	}

	fmt.Println("Last changed:", cache.LastChanged)
	cache.LastChanged = time.Now().Unix()
	return true, nil
}

func syncData() {
	for {
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

func clearCachedUser(userid int64) {
	for id, member := range cache.Member {
		if member.Id == userid {
			cache.Member = append(cache.Member[:id], cache.Member[id+1:]...)
		}
	}
}

func clearDeleteListByUser(userid int64) {
	for id, message := range cache.DeleteList {
		if message.UserID == userid {
			cache.DeleteList = append(cache.DeleteList[:id], cache.DeleteList[id+1:]...)
		}
	}
}
