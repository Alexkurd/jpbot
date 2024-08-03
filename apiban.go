package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

var CASBAN_API = "https://api.cas.chat/check?user_id="
var LOLSBOT_API = "https://api.lols.bot/account?id="

type casban_response struct {
	Status      bool   `json:"ok"`
	Description string `json:"description"`
}

var casban casban_response

type lolsbot_response struct {
	Status   bool    `json:"banned"`
	Offenses int     `json:"offenses"`
	Score    float32 `json:"spam_factor"`
}

var lolsbot lolsbot_response

func isUserApiBanned(userid int) bool {
	casbanned := isUserCasBanned(userid)
	lolsbanned := isUserLolsBanned(userid)
	log.Printf("CAS: %t", casbanned)
	log.Printf("LOLS %t", lolsbanned)
	return casbanned || lolsbanned
}

func isUserCasBanned(userid int) bool {
	// Send GET request
	resp, err := http.Get(CASBAN_API + fmt.Sprint(userid))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}

	defer resp.Body.Close()
	// Check for successful response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.StatusCode)
		return false
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	err = json.Unmarshal(body, &casban)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return false
	}
	if casban.Status {
		log.Print("User " + strconv.Itoa(userid) + " is CasBanned")
		log.Print(casban)
	}
	return casban.Status
}

func isUserLolsBanned(userid int) bool {
	// Send GET request
	resp, err := http.Get(LOLSBOT_API + fmt.Sprint(userid))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}

	defer resp.Body.Close()
	// Check for successful response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.StatusCode)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	err = json.Unmarshal(body, &lolsbot)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return false
	}

	if lolsbot.Status {
		log.Print("User " + strconv.Itoa(userid) + " is LolsBanned")
	}

	return lolsbot.Status
}
