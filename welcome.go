package main

/*
 - Show welcome message on join with inline-button.
 - Set lowest rights
 - If button clicked, set guest rights, remove welcome message.
*/

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	var keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(MainConfig.WelcomeButtonMessage, "upgrade_rights"),
		),
	)

	msg.DisableWebPagePreview = true
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	if emulate {
		log.Println("Welcome for " + user.UserName)
	} else {
		bot.Send(msg)
	}
}

func setInitialRights(update tgbotapi.Update, user tgbotapi.User) {
	// Example: Set user rights to read-only initially
	if emulate {
		log.Print("Setting rights for user: ", user.ID, user.UserName)
		return
	}

	var chatid int64
	if update.Message == nil {
		chatid = update.ChatMember.Chat.ID
	} else {
		chatid = update.Message.Chat.ID
	}

	initialRights := tgbotapi.ChatPermissions{
		CanSendMessages:       false,
		CanSendMediaMessages:  false,
		CanSendOtherMessages:  false,
		CanAddWebPagePreviews: false,
	}

	config := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatid,
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
	switch query.Data {
	case "upgrade_rights":
		upgradeUserRights(query)
		answerCallbackQuery(query.ID, "Rights upgraded!")
		deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
		// handle other callbacks here
	}
}

func upgradeUserRights(query *tgbotapi.CallbackQuery) {
	// Logic to upgrade user rights
	// Example: giving the user the ability to send messages
	defaultRights := tgbotapi.ChatPermissions{
		CanSendMessages: true,
	}

	config := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: query.Message.Chat.ID,
			UserID: query.From.ID,
		},
		Permissions: &defaultRights,
	}
	bot.Request(config)
}

func answerCallbackQuery(callbackQueryID string, text string) {
	callbackConfig := tgbotapi.NewCallback(callbackQueryID, text)
	bot.Send(callbackConfig)
}
