package main

/*
 - Show welcome message on join with inline-button.
 - Set lowest rights
 - If button clicked, set guest rights, remove welcome message.
*/

import (
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/Alexkurd/telegram-bot-api/v7"
)

type WelcomeMessage struct {
	ID        int
	UserID    int64
	ChatID    int64
	Timestamp time.Time
}

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
		messageSent, _ := bot.Send(msg)
		welcomeSent(messageSent, user.ID)
		log.Println("Welcome sent ", messageSent.MessageID, " user "+user.UserName, user.ID)
		saveCache()
	}
}

func setInitialRights(update tgbotapi.Update, user tgbotapi.User) {
	log.Print("Setting rights for user: ", user.ID, " ", user.UserName)
	//Set user rights to read-only initially
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

func upgradeUserRights(chatID int64, userid int64) {
	// Logic to upgrade user rights
	// Giving the user the ability to send messages
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
	clearCachedUser(userid)
	clearDeleteListByUser(userid)
}

func answerCallbackQuery(callbackQueryID string, text string) {
	callbackConfig := tgbotapi.NewCallback(callbackQueryID, text)
	bot.Send(callbackConfig)
}

func welcomeSent(message tgbotapi.Message, userID int64) {
	// Add message with timestamp + 2 hours
	cache.DeleteList = append(cache.DeleteList, WelcomeMessage{
		ID:        message.MessageID,
		UserID:    userID,
		ChatID:    message.Chat.ID,
		Timestamp: time.Now().UTC().Add(time.Hour * 2),
	})
}

func CleanUpWelcome() int {
	// Remove entries older than now
	now := time.Now().UTC()
	counter := 0
	if len(cache.DeleteList) > 0 {
		for id := 0; id < len(cache.DeleteList); id++ {
			if cache.DeleteList[id].Timestamp.Before(now) {
				log.Println("Deleting message id", cache.DeleteList[id].ID)
				deleteMessage(cache.DeleteList[id].ChatID, cache.DeleteList[id].ID)
				kickChatMember(cache.DeleteList[id].ChatID, cache.DeleteList[id].UserID)
				clearCachedUser(cache.DeleteList[id].UserID)
				//clearDeleteListByUser(message.UserID)
				cache.DeleteList = append(cache.DeleteList[:id], cache.DeleteList[id+1:]...)
				id--
				counter++
			}
		}
		saveCache()
	}
	return counter
}
