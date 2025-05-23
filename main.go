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
	DenyChats            []string          `yaml:"denychats"`
	DenyNames            []string          `yaml:"denynames"`
	Admins               []int             `yaml:"admins"`
	PinnedMessage        string            `yaml:"pinnedMessage"`
	PinnedMessageId      int               `yaml:"pinnedMessageId"`
}

const TEXTMESSAGE_LIMIT = 4096

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
	go initMetrics()
	updates = startBot()
	processUpdates(updates)

} //End main()

func processUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		requestsTotal.Inc()
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
			if update.ChatMember.NewChatMember.Status == "left" {
				continue
			}
			if update.ChatMember.NewChatMember.Status == "restricted" {
				continue
			}

			if isBadName(update.ChatMember) {
				BanChatMember(update.ChatMember.Chat.ID, update.ChatMember.NewChatMember.User.ID, 0)
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
			if update.EditedMessage != nil {
				CheckTriggerMessage(update.EditedMessage)
			}
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			processCommands(update.Message.Command(), *update.Message)
			continue
		}

		// Handle new members joining
		if update.Message.NewChatMembers != nil {
			//Clean old triggers
			cleanTriggers()
			//Check members
			for _, newMember := range update.Message.NewChatMembers {
				if isCachedUser(newMember.ID, update.FromChat().ID) {
					welcomeNewUser(update, newMember)
					setInitialRights(update, newMember)
					checkCachedQueue()
					continue
				}
			}
		}

		//Handle member left
		if update.Message.LeftChatMember != nil {
			slog.Info("Member left: " + update.Message.LeftChatMember.UserName)
			slog.Info("Update.Message" + update.Message.Text)
			//log.Print(fmt.Printf("%+v\n", update.Message))
			continue
		}

		if isDenyBot(update.Message) {
			deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			continue
		}

		if isDenyChat(update.Message) {
			deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			continue
		}

		// Check for forbidden text
		if !isAdmin(update.Message.From.ID) {
			// Check Bad text message
			if isBadMessage(update.Message.Text) {
				if emulate {
					log.Print(update.Message.Chat.ID, update.Message.MessageID)
					continue
				}
				deleteMessage(update.Message.Chat.ID, update.Message.MessageID)
				CleanUpWelcome()
				//CleanWelcomeQueue()
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
		} else {
			//AdminsZone
			if update.Message.ChatShared != nil {
				if update.Message.ChatShared.RequestID == 1000 { //pin message
					pinMessage(update.Message.ChatShared.ChatID)
				}
			}
		}

		CheckTriggerMessage(update.Message)

		//Fix rights for the newcomers
		fixRights(update)

		collectMapUrls(*update.Message)
	}
}

func pinMessage(id int64) {
	msg := tgbotapi.NewMessage(id, getPinnedMessage())
	msg.ParseMode = "HTML"
	mId, _ := bot.Send(msg)
	slog.Info(fmt.Sprintf("Pinned message ID: %d", mId.MessageID))
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
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Дополнительные команды в кнопке меню.\r\n Посмотрите готовые статьи")
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

	if member.CanSendMessages && !member.CanAddWebPagePreviews {
		slog.Info("Fix rights for user " + update.Message.From.UserName)
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
	case "unban":
		if isAdmin(message.From.ID) {
			//TODO as it's hard to get userId
		}
	case "uptime":
		msg.Text = "Uptime: " + uptime()
		msg.ReplyParameters.MessageID = message.MessageID
	case "start":
		msg.Text = "Дополнительные команды в кнопке меню.\r\nПосмотрите готовые статьи"
		msg.ReplyMarkup = rootMenu()
		deleteMessage(message.Chat.ID, message.MessageID)
	case "reload":
		reload()
		msg.Text = "Reloaded"
	case "triggers":
		msg.ParseMode = "HTML"
		msg.Text = getTriggersList()
	case "clean_triggers":
		counter := cleanTriggers()
		msg.Text = "Cleaned " + strconv.Itoa(counter) + " trigger messages"
	case "clean_welcome":
		counter := checkCachedQueue()
		msg.Text = "Cleaned " + strconv.Itoa(counter) + " welcomed users"
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
		msg.Text = toggleDebugmode()
	case "force_mode":
		msg.Text = toggleForcemode()
	default:
		msg.Text = ""
	}
	if msg.Text != "" {
		if len(msg.Text) < TEXTMESSAGE_LIMIT {
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else {
			messages := strings.Split(msg.Text, "\r\n\r\n")
			for _, message := range messages {
				partialMessage := msg
				partialMessage.Text = message
				_, err = bot.Send(partialMessage)
			}
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

func checkCachedQueue() int {
	counter := 0
	if len(cache.Member) > 0 {
		for id := 0; id < min(20, len(cache.Member)); id++ {
			if isUserApiBanned(int(cache.Member[id].Id)) {
				if cache.Member[id].ChatId == 0 {
					cache.Member[id].ChatId = -1001164690983
				}
				BanChatMember(cache.Member[id].ChatId, cache.Member[id].Id, time.Now().Unix()+10)
				unbanChatMember(cache.Member[id].ChatId, cache.Member[id].Id)
				cache.Member = append(cache.Member[:id], cache.Member[id+1:]...)
				counter++
			}
		}
	}
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
	admins, _ := bot.GetChatAdministrators(tgbotapi.ChatAdministratorsConfig{ChatConfig: tgbotapi.ChatConfig{
		ChannelUsername: "@wrenjapanchat", //TODO Move to .env
	}})

	for _, admin := range admins {
		MainConfig.Admins = append(MainConfig.Admins, int(admin.User.ID))
	}
	MainConfig.Admins = unique(MainConfig.Admins)

	slog.Info(fmt.Sprintf("Admins: %v", MainConfig.Admins))

	//
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
		u.Timeout = 7
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
