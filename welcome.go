package main

/*
 - Show welcome message on join with inline-button.
 - Set lowest rights
 - If button clicked, set guest rights, remove welcome message.
*/

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/Alexkurd/telegram-bot-api/v7"
)

func welcomeNewUser(update tgbotapi.Update, user tgbotapi.User) {
	var chatid int64
	if update.Message == nil {
		chatid = update.ChatMember.Chat.ID
	} else {
		chatid = update.Message.Chat.ID
	}

	// Greet new user
	welcomeMessage := strings.Replace(MainConfig.WelcomeMessage, "{namelink}", getNameLink(user), -1)
	msg := tgbotapi.NewMessage(chatid, welcomeMessage)

	//Add Inline callback
	userid := strconv.Itoa(int(user.ID))
	callbackData := "{\"command\": \"upgrade_rights\", \"data\": \"" + userid + "\"}"

	var keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(MainConfig.WelcomeButtonMessage, callbackData),
		),
	)
	msg.LinkPreviewOptions.IsDisabled = true
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	if emulate {
		log.Println("Welcome for " + user.UserName)
	} else {
		bot.Send(msg)
	}
}

func setInitialRights(update tgbotapi.Update, user tgbotapi.User) {
	log.Print("Setting rights for user: ", user.ID, user.UserName)
	// Example: Set user rights to read-only initially
	if emulate {
		return
	}

	var chatid int64
	if update.Message == nil {
		chatid = update.ChatMember.Chat.ID
	} else {
		chatid = update.Message.Chat.ID
	}

	initialRights := tgbotapi.ChatPermissions{
		CanSendMessages: false,
	}

	config := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: chatid,
			},
			UserID: user.ID,
		},
		Permissions: &initialRights,
	}
	_, err := bot.Request(config)
	if err != nil {
		log.Print(err)
	}
}

func handleCallback(query *tgbotapi.CallbackQuery) {
	type Command struct {
		Command string `json:"command"`
		Data    string `json:"data"`
	}
	var callback Command
	err := json.Unmarshal([]byte(query.Data), &callback)
	if err != nil {
		log.Print(err)
	}
	switch callback.Command {
	case "upgrade_rights":
		if callback.Data != strconv.Itoa(int(query.From.ID)) {
			log.Print("User " + query.From.UserName + ":" + strconv.Itoa(int(query.From.ID)) + " clicked wrong button")
			break
		}
		upgradeUserRights(query.Message.Chat.ID, query.From.ID)
		//member := tgbotapi.SetChatM{

		answerCallbackQuery(query.ID, "Rights upgraded!")
		deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
		// handle other callbacks here
	}
}

func upgradeUserRights(chatID int64, userid int64) {
	// Logic to upgrade user rights
	// Example: giving the user the ability to send messages
	defaultRights := tgbotapi.ChatPermissions{
		CanSendMessages: true,
	}
	defaultRights.SetCanSendMediaMessages(true)

	config := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: chatID,
			},
			UserID: userid,
		},
		UseIndependentChatPermissions: true,
		Permissions:                   &defaultRights,
	}
	bot.Send(config)
}

func answerCallbackQuery(callbackQueryID string, text string) {
	callbackConfig := tgbotapi.NewCallback(callbackQueryID, text)
	bot.Send(callbackConfig)
}
