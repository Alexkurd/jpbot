package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token                string            `yaml:"bot_token"`
	Connection           string            `yaml:"connection"`
	HostPort             string            `yaml:"hostport"`
	ForbiddenText        []string          `yaml:"forbiddenText"`
	Ranks                map[string]string `yaml:"ranks"`
	WelcomeMessage       string            `yaml:"welcome_message"`
	WelcomeButtonMessage string            `yaml:"welcome_button_message"`
	DenyBots             []string          `yaml:"denybots"`
	Admins               []int             `yaml:"admins"`
}

var emulate = false
var forceProtection = true

var MainConfig Config
var startTime time.Time
var bot *tgbotapi.BotAPI
var err error

func init() {
	readConfig() //Fill config with values
	readTriggers()
	importCache()
	startTime = time.Now()
	go syncData()
}

func main() {
	var updates tgbotapi.UpdatesChannel
	botInit()
	afterBotInit()
	//	initMetrics()
	updates = startBot()
	processUpdates(updates)
} //End main()

func processUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Check for callback query
		if update.CallbackQuery != nil {
			handleCallback(update.CallbackQuery)
			continue
		}

		//Private chat
		if update.MyChatMember != nil {
			slog.Info(fmt.Sprintf("New MyChatMember %d", update.MyChatMember.NewChatMember.User.ID))
		}
		//If chat hides userlist
		if update.ChatMember != nil {
			if update.ChatMember.NewChatMember.Status == "kicked" {
				continue
			}
			if isNewMember(update.ChatMember) {
				setInitialRights(update, *update.ChatMember.NewChatMember.User)
				if forceProtection {
					if isUserApiBanned(int(update.ChatMember.NewChatMember.User.ID)) {
						BanChatMember(update.ChatMember.Chat.ID, update.ChatMember.NewChatMember.User.ID, 0)
					} else {
						welcomeNewUser(update, *update.ChatMember.NewChatMember.User)
					}
				} else {
					welcomeNewUser(update, *update.ChatMember.NewChatMember.User)
				}
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
				if isCachedUser(newMember.ID) {
					welcomeNewUser(update, newMember)
					setInitialRights(update, newMember)
				}
			}
		}

		//Handle member left
		if update.Message.LeftChatMember != nil {
			slog.Info("Member left: " + update.Message.LeftChatMember.UserName)
			slog.Info("Update.Message" + update.Message.Text)
			//log.Print(fmt.Printf("%+v\n", update.Message))
		}

		if isDenyBot(update.Message) {
			deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
		}

		// Check for forbidden text
		if !isAdmin(update.Message.Chat.ID) {
			// Check Bad text message
			if isBadMessage(update.Message.Text) {
				if emulate {
					log.Print(update.Message.Chat.ID, update.Message.MessageID)
					continue
				}
				deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
				CleanUpWelcome()
				continue
			}
			//Check message starting with emoji. Usually spam.
			if isMessageStartsWithEmoji(update) {
				slog.Info("Deleted message with emoji - " + update.Message.From.UserName)
				deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			}
			//Check message from channel
			if isChannelMessage(update) {
				slog.Info("Deleted message from channel - " + update.Message.SenderChat.UserName)
				deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
				continue
			}
		}

		CheckTriggerMessage(update.Message)
		if isAdmin(update.Message.From.ID) {
			slog.Info("New Admin from " + update.Message.From.UserName)
		}

		//Fix rights for the newcomers
		fixRights(update)
	}
}

func handleCallback(query *tgbotapi.CallbackQuery) {
	type Command struct {
		Command string `json:"command"`
		Data    string `json:"data"`
	}
	var callback Command
	var user int64
	err := json.Unmarshal([]byte(query.Data), &callback)
	if err != nil {
		slog.Warn("Callback error:", "error", err)
	}
	switch callback.Command {
	case "upgrade_rights":
		user, err = strconv.ParseInt(callback.Data, 10, 64)
		if !isAdmin(query.From.ID) {
			if user != query.From.ID {
				slog.Info(fmt.Sprintf("User %s(%d) clicked wrong button", query.From.UserName, query.From.ID))
				break
			}
		}
		slog.Info(fmt.Sprintf("User %s(%d) clicked his button", query.From.UserName, query.From.ID))
		if isUserApiBanned(int(user)) {
			answerCallbackQuery(query.ID, "Sorry, Api Ban")
			BanChatMember(query.Message.Chat.ID, user, 0)
		} else {
			upgradeUserRights(query.Message.Chat.ID, user)
			answerCallbackQuery(query.ID, "Rights upgraded!")
		}
		deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
	// handle other callbacks here
	case "show_menu":
		deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "ᓚᘏᗢ"+strings.Repeat(" ", 80)+"\\^o^/")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = replyWithMenu(callback.Data)
		bot.Send(msg)
	case "show_root_menu":
		deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Посмотрите готовые статьи")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = rootMenu()
		bot.Send(msg)
	}
}

func fixRights(update tgbotapi.Update) {
	config := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatConfig: tgbotapi.ChatConfig{
				ChatID: update.Message.Chat.ID,
			},
			UserID: update.Message.From.ID,
		},
	}

	member, err := bot.GetChatMember(config)
	if err != nil {
		log.Print(err)
	}

	if member.CanSendMessages && !member.CanSendPhotos {
		log.Print("Fix rights for user " + update.Message.From.UserName)
		upgradeUserRights(update.Message.Chat.ID, update.Message.From.ID)
	}
}

func processCommands(command string, message tgbotapi.Message) {
	//Only Private messages
	if message.From.ID != message.Chat.ID {
		deleteMessage(message.Chat.ID, message.MessageID)
		return
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch command {
	case "help":
		msg.Text = "I understand /uptime and /start."
		msg.ReplyParameters.MessageID = message.MessageID
	case "uptime":
		msg.Text = "Uptime: " + uptime()
		msg.ReplyParameters.MessageID = message.MessageID
	case "start":
		msg.Text = "Посмотрите готовые статьи"
		msg.ReplyMarkup = rootMenu()
		deleteMessage(message.Chat.ID, message.MessageID)
	case "reload":
		reload()
		msg.Text = "Reloaded"
	case "triggers":
		msg.ParseMode = "HTML"
		msg.Text = getTriggersList()
	case "say":
		msg.Text = "Select chat to send"
		msg.ReplyMarkup = say()
	case "deletequeue":
		msg.Text = ToDeleteQueue()
	case "checkqueue":
		counter := checkBanQueue()
		msg.Text = "Cleaned " + strconv.Itoa(counter) + " messages"
	case "welcomequeue":
		CleanWelcomeQueue()
		msg.Text = "Welcome queue cleaned"
	case "cleanup":
		counter := CleanUpWelcome()
		msg.Text = "Cleaned " + strconv.Itoa(counter) + " messages"
	case "debug_mode":
		msg.Text = toggle_debugMode()
	case "force_mode":
		msg.Text = toggle_forceMode()
	default:
		msg.Text = ""
	}
	if msg.Text != "" {
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func say() tgbotapi.ReplyKeyboardMarkup {
	var buttonRow []tgbotapi.KeyboardButton

	buttonRow = append(buttonRow,
		tgbotapi.KeyboardButton{
			RequestChat: &tgbotapi.KeyboardButtonRequestChat{RequestID: 1000, BotIsMember: true},
			Text:        "Select chat",
		})

	return tgbotapi.NewOneTimeReplyKeyboard(buttonRow)
}

func ToDeleteQueue() string {
	return fmt.Sprintln(cache.DeleteList)
}

func CleanWelcomeQueue() {
	cache.Member = nil
}

func checkBanQueue() int {
	counter := 0
	if len(cache.DeleteList) > 0 {
		for id := 0; id < len(cache.DeleteList); id++ {
			if isUserApiBanned(int(cache.DeleteList[id].UserID)) {
				cache.DeleteList[id].Timestamp = time.Now().UTC()
				counter++
			}
		}
	}
	CleanUpWelcome()
	return counter
}

func uptime() string {
	return fmt.Sprintln(time.Since(startTime).Round(time.Second))
}

func reload() {
	readTriggers()
	readMenu()
}

func readConfig() {
	// Read and parse the config file
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Panic(err)
	}

	err = yaml.Unmarshal(configFile, &MainConfig)
	if err != nil {
		log.Panic(err)
	}
	//Override with ENV
	token := os.Getenv("BOT_TOKEN")
	if token != "" {
		MainConfig.Token = token
	}
	hostport := os.Getenv("WEBHOOK_HOSTPORT")
	if token != "" {
		MainConfig.HostPort = hostport
	}
}

func botInit() {
	//Common part
	bot, err = tgbotapi.NewBotAPI(MainConfig.Token) // Set up the Telegram bot
	if err != nil {
		log.Panic(err)
	}
	if isDebugMode() {
		bot.Debug = true
	}

	slog.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))
}

func afterBotInit() {
	startMenu()
}

func startBot() tgbotapi.UpdatesChannel {
	var updates tgbotapi.UpdatesChannel

	if !isWebhook() {
		deleteWh := tgbotapi.DeleteWebhookConfig{}
		bot.Request(deleteWh)
		//GetUpdatesWay
		// Create a new Update config with a timeout
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 10
		updates = bot.GetUpdatesChan(u)

		// TODO Ranking
		// Initialize a map to track user message counts
		//userMessageCounts := make(map[int]int)
		//End GetUpdates

		//Webhook way
	} else {
		wh, err := tgbotapi.NewWebhookWithCert("https://"+MainConfig.HostPort+"/"+bot.Token, tgbotapi.FilePath("jpbot.pem"))
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

		updates = bot.ListenForWebhook("/" + bot.Token)

		go http.ListenAndServeTLS(MainConfig.HostPort, "jpbot.pem", "jpbot.key", nil)
	}

	return updates
}
