package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var CASBAN_API = "https://api.cas.chat/check?user_id="

type casban_response struct {
	Status      bool   `json:"ok"`
	Description string `json:"description"`
}

var casban casban_response

func isUserCasBanned(userid int) bool {
	// Send GET request
	resp, err := http.Get(CASBAN_API + fmt.Sprint(userid))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}

	defer resp.Body.Close()

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

	return casban.Status
}
