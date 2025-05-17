package main

import (
	"encoding/csv"
	"fmt"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"log/slog"
	"os"
	"regexp"
	"strconv"
)

func collectMapUrls(message tgbotapi.Message) {
	urlPattern := `https://maps\.app\.goo\.gl/[^\s]+`
	filename := "maps.csv"
	re := regexp.MustCompile(urlPattern)

	// Find the first matching URL in the message
	fullURL := re.FindString(message.Text)
	if fullURL != "" {
		slog.Info("Detected new map URL: " + fullURL)
		// Check if the file exists
		err := createCSVFile(filename)
		if err != nil {
			return
		}
		// Open or create the CSV file
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.Error("Failed to open or create file: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		err = writer.Write([]string{message.Time().UTC().String(), message.Text, fullURL, message.From.UserName, strconv.Itoa(int(message.From.ID)), strconv.FormatBool(message.Chat.IsPrivate()), "https://t.me/" + message.Chat.UserName + "/" + strconv.Itoa(message.MessageID)})
		if err != nil {
			slog.Error("Failed to write to file: %w", err)
		}
		writer.Flush()
	}
}

func createCSVFile(filename string) error {
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the file since it does not exist
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		// Write UTF-8 BOM to the file
		bom := []byte{0xEF, 0xBB, 0xBF} // UTF-8 BOM bytes
		if _, err := file.Write(bom); err != nil {
			return fmt.Errorf("failed to write BOM to file: %w", err)
		}
	}
	return nil
}
