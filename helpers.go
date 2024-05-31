package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/Alexkurd/telegram-bot-api/v7"
)

func isWebhook() bool {
	return MainConfig.Connection == "webhook"
}

func isDebugMode() bool {
	return strings.ToLower(os.Getenv("BOT_DEBUG")) == "true"
}

func isCachedUser(userid int64) bool {
	newMember := getMember(userid)
	if newMember == nil {
		cache.Member = append(cache.Member, ChatMember{
			Id:            userid,
			WelcomeShowed: true,
			Rank:          0,
			MessageCount:  0,
		})
		return true
	}
	return false
}

func isNewMember(Member *tgbotapi.ChatMemberUpdated) bool {
	if Member.OldChatMember.IsMember && !Member.NewChatMember.IsMember {
		log.Print("IsMember state changed ", Member.OldChatMember.IsMember, "->", Member.NewChatMember.IsMember)
		return false
	}

	return isCachedUser(Member.NewChatMember.User.ID)
}

func isMessageStartsWithEmoji(update tgbotapi.Update) bool {
	if update.Message.Entities == nil {
		return false
	}
	if update.Message.Entities[0].Type == "custom_emoji" && update.Message.Entities[0].Offset == 0 && len(update.Message.Text) > 4 {
		return true
	}
	return false
}

func kickChatMember(chatID int64, userID int64) {
	BanChatMember(chatID, userID, time.Now().UTC().Add(time.Hour*6).Unix())
}

func BanChatMember(chatID int64, userID int64, untilDate int64) {
	//Ban for 11 months
	if untilDate == 0 {
		untilDate = time.Now().UTC().AddDate(0, 11, 0).Unix()
	}

	config := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: chatID,
			},
			UserID: userID,
		},
		UntilDate: untilDate,
	}
	bot.Send(config)
	log.Print("User banned: ", userID, " until ", untilDate)
}

func isChannelMessage(update tgbotapi.Update) bool {
	return update.Message.SenderChat != nil
}

func isDenyBot(message *tgbotapi.Message) bool {
	badbot := false
	if message.ViaBot != nil {
		for _, bot := range MainConfig.DenyBots {
			if strings.ToLower(message.ViaBot.UserName) == bot {
				badbot = true
				break
			}
		}
	}
	return badbot
}

func getNameLink(user tgbotapi.User) string {
	log.Print(user)
	userid := strconv.Itoa(int(user.ID))
	name := ""
	if user.FirstName != "" {
		name = user.FirstName
	}
	if user.LastName != "" {
		name = name + " " + user.LastName
	}
	if user.UserName != "" {
		name = name + "(" + user.UserName + ")"
	}
	return "<a href=\"tg://user?id=" + userid + "\">" + name + "</a>"
}

func deleteMessage(chatID int64, messageId int) {
	deleteConfig := tgbotapi.DeleteMessageConfig{
		BaseChatMessage: tgbotapi.BaseChatMessage{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: chatID,
			},
			MessageID: messageId,
		},
	}
	bot.Send(deleteConfig)
}

func isBadMessage(message string) bool {
	for _, word := range MainConfig.ForbiddenText {
		if word[0] == 'r' {
			regex := regexp.MustCompile(word[2:])
			if regex.MatchString(message) {
				log.Print("TriggeredBad: ", word[2:])
				return true
			}
		} else {
			if strings.Contains(message, word) {
				log.Print("TriggeredBad: ", word)
				return true
			}
		}
	}
	return false
}
