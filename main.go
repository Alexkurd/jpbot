package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token                string            `yaml:"bot_token"`
	Connection           string            `yaml:"connection"`
	ForbiddenText        []string          `yaml:"forbiddenText"`
	Ranks                map[string]string `yaml:"ranks"`
	WelcomeMessage       string            `yaml:"welcome_message"`
	WelcomeButtonMessage string            `yaml:"welcome_button_message"`
}

var emulate = false

var MainConfig Config
var startTime time.Time
var bot *tgbotapi.BotAPI

func readConfig() {
	// Read and parse the config file
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Panic(err)
	}
	//var config Config
	err = yaml.Unmarshal(configFile, &MainConfig)
	if err != nil {
		log.Panic(err)
	}
	//Override with ENV
	token := os.Getenv("BOT_TOKEN")
	if token != "" {
		MainConfig.Token = token
	}
}

func main() {
	var err error

	readConfig() //Fill config with values
	readTriggers()
	importData()
	startTime = time.Now()
	go syncData()
	//Common part
	bot, err = tgbotapi.NewBotAPI(MainConfig.Token) // Set up the Telegram bot
	if err != nil {
		log.Panic(err)
	}
	if isDebugMode() {
		bot.Debug = true
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	if !isWebhook() {
		deleteWh := tgbotapi.DeleteWebhookConfig{}
		bot.Request(deleteWh)
		//GetUpdatesWay
		// Create a new Update config with a timeout
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 10
		updates := bot.GetUpdatesChan(u)

		// TODO Ranking
		// Initialize a map to track user message counts
		//userMessageCounts := make(map[int]int)
		processUpdates(updates)
		//End GetUpdates

		//Webhook way
	} else {
		wh, err := tgbotapi.NewWebhookWithCert("https://95.217.5.60:8443/"+bot.Token, tgbotapi.FilePath("jpbot.pem"))

		if err != nil {
			panic(err)
		}

		_, err = bot.Request(wh)

		if err != nil {
			panic(err)
		}

		info, err := bot.GetWebhookInfo()

		if err != nil {
			panic(err)
		}

		if info.LastErrorDate != 0 {
			log.Printf("failed to set webhook: %s", info.LastErrorMessage)
		}
		updates := bot.ListenForWebhook("/" + bot.Token)

		go http.ListenAndServeTLS("95.217.5.60:8443", "jpbot.pem", "jpbot.key", nil)

		// for update := range updates {
		// 	log.Printf("%+v\n", update)
		// }

		// updates := bot.ListenForWebhook("/" + bot.Token)
		// http.ListenAndServeTLS("95.217.5.60:88", "jpbot.pem", "jpbot.key", nil)
		processUpdates(updates)
	}

} //End main()

func processUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Check for callback query
		if update.CallbackQuery != nil {
			handleCallback(update.CallbackQuery)
		}
		//Custom_emoji
		//update.Message.CaptionEntities[].Type == "custom_emoji"

		//Private chat
		if update.MyChatMember != nil {
			log.Print("New MyChatMember", update.MyChatMember.NewChatMember.User.ID)
		}
		//If chat hides userlist
		if update.ChatMember != nil {
			log.Print("UpdateId: ", update.UpdateID)
			log.Print("New ChatMember joined " + update.ChatMember.NewChatMember.User.UserName)
			if isNewMember(update.ChatMember.NewChatMember.User.ID) {
				welcomeNewUser(update, *update.ChatMember.NewChatMember.User)
				setInitialRights(update, *update.ChatMember.NewChatMember.User)
			}
		}

		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			processCommands(update.Message.Command(), *update.Message)
			continue
		}

		// Handle new members joining
		if update.Message.NewChatMembers != nil {
			for _, newMember := range update.Message.NewChatMembers {
				log.Print("New user joined " + newMember.UserName)
				if isNewMember(newMember.ID) {
					welcomeNewUser(update, newMember)
					setInitialRights(update, newMember)
				}
			}
		}

		// Check for forbidden text
		if isBadMessage(update.Message.Text) {
			if emulate {
				log.Print(update.Message.Chat.ID, update.Message.MessageID)
				continue
			}
			deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			continue
		}

		if update.Message.ReplyToMessage != nil {
			CheckTriggerMessage(update.Message)
		}
	}
}

func getNameLink(user tgbotapi.User) string {
	log.Print(user)
	userid := strconv.Itoa(int(user.ID))
	name := ""
	if user.FirstName != "" {
		name = user.FirstName
	}
	if user.LastName != "" {
		name = name + "" + user.LastName
	}
	return "<a href=\"tg://user?id=" + userid + "\">" + name + "</a>"
}

func deleteMessage(chatID int64, messageId int) {
	deleteConfig := tgbotapi.DeleteMessageConfig{
		MessageID: messageId,
		ChatID:    chatID,
	}
	bot.Send(deleteConfig)
}

func isBadMessage(message string) bool {
	for _, word := range MainConfig.ForbiddenText {
		if word[0] == 'r' {
			regex := regexp.MustCompile(word[2:])
			if regex.MatchString(message) {
				return true
			}
		} else {
			if strings.Contains(message, word) {
				return true
			}
		}
	}
	return false
}

func processCommands(command string, message tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.ReplyToMessageID = message.MessageID
	switch command {
	case "help":
		msg.Text = "I understand /uptime and /status."
	case "uptime":
		msg.Text = "Uptime: " + uptime()
	case "status":
		msg.Text = "I'm ok."
	default:
		msg.Text = ""
	}
	if msg.Text != "" {
		bot.Send(msg)
	}
}

func uptime() string {
	return fmt.Sprintln(time.Since(startTime).Round(time.Second))
}

func isWebhook() bool {
	return MainConfig.Connection == "webhook"
}

func isDebugMode() bool {
	return strings.ToLower(os.Getenv("BOT_DEBUG")) == "true"
}

func isNewMember(userid int64) bool {
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
