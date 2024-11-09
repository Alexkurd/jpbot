package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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
	slog.Info(fmt.Sprintf("User %d ban status: CAS:%t LOLS:%t", userid, casbanned, lolsbanned))
	return casbanned || lolsbanned
}

func isUserCasBanned(userid int) bool {
	http.DefaultClient.Timeout = 10 * time.Second
	// Send GET request
	resp, err := http.Get(CASBAN_API + fmt.Sprint(userid))
	if err != nil {
		slog.Warn("Error sending request:", "error", err)
		return false
	}

	defer resp.Body.Close()
	// Check for successful response status code
	if resp.StatusCode != http.StatusOK {
		slog.Warn("Error status code:", "statusCode", resp.StatusCode)
		return false
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("Error reading response body:", "error", err)
		return false
	}

	err = json.Unmarshal(body, &casban)
	if err != nil {
		slog.Warn("Error unmarshalling JSON:", "error", err)
		return false
	}
	if casban.Status {
		slog.Debug("User "+strconv.Itoa(userid)+" is CasBanned", "response", casban)
	}
	return casban.Status
}

func isUserLolsBanned(userid int) bool {
	http.DefaultClient.Timeout = 10 * time.Second
	// Send GET request
	resp, err := http.Get(LOLSBOT_API + fmt.Sprint(userid))
	if err != nil {
		slog.Warn("Error sending request:", "error", err)
		return false
	}

	defer resp.Body.Close()
	// Check for successful response status code
	if resp.StatusCode != http.StatusOK {
		slog.Warn("Error status code:", "statusCode", resp.StatusCode)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("Error reading response body:", "error", err)
		return false
	}

	err = json.Unmarshal(body, &lolsbot)
	if err != nil {
		slog.Warn("Error unmarshalling JSON:", "error", err)
		return false
	}

	if lolsbot.Status {
		slog.Debug("User "+strconv.Itoa(userid)+" is CasBanned", "response", lolsbot)
	}

	return lolsbot.Status
}
