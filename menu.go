package main

import (
	"log"
	"os"

	tgbotapi "github.com/Alexkurd/telegram-bot-api/v7"
	"gopkg.in/yaml.v3"
)

type Menu struct {
	Title    string `yaml:"Title"`
	Command  string `yaml:"Command"`
	Submenu  []Menu `yaml:"Submenu"`
	Flatmenu map[string]Menu
	Url      string `yaml:"Url"`
}

var menu []Menu
var flatMenu map[string]Menu

func startMenu() {
	readMenu()

	// Register command handlers
	registerCommandHandlers()
}

func readMenu() {
	// Parse menu.yaml
	data, err := os.ReadFile("menu.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(data, &menu)
	if err != nil {
		log.Fatal(err)
	}

	flatMenu = flattenMenu(menu)
}

func flattenMenu(menu []Menu) map[string]Menu {
	flatItems := make(map[string]Menu)

	for _, item := range menu {
		fullCommand := "." + item.Command
		var flatmenu map[string]Menu

		if item.Submenu != nil {
			flatmenu = flattenMenu(item.Submenu)
		}

		flatItems[fullCommand] = Menu{
			Title:    item.Title,
			Command:  fullCommand,
			Submenu:  item.Submenu,
			Flatmenu: flatmenu,
			Url:      item.Url,
		}
	}
	return flatItems
}

// Add commands for Mennu button
func registerCommandHandlers() {

	//Generate commands for Menu button
	var commands []tgbotapi.BotCommand

	commands = append(commands, tgbotapi.BotCommand{
		Command:     "/start",
		Description: "Показать главное меню",
	},
		tgbotapi.BotCommand{
			Command:     "/triggers",
			Description: "Список быстрых команд",
		},
	)

	/*for _, item := range menu {
		commands = append(commands, tgbotapi.BotCommand{
			Command:     item.Command,
			Description: item.Text,
		})
	}*/

	//Set menu for Private chats
	scope := tgbotapi.NewBotCommandScopeAllPrivateChats()
	menuConfig := tgbotapi.SetChatMenuButtonConfig{
		MenuButton: &tgbotapi.MenuButton{
			Type: "commands",
		},
	}
	bot.Send(tgbotapi.NewSetMyCommandsWithScope(scope, commands...))
	bot.Send(menuConfig)
}

func rootMenu() tgbotapi.InlineKeyboardMarkup {
	return generateMenuKeyboard(menu, false)
}

func prepareCallbackMenuCommand(command string) string {
	return "{\"command\": \"show_menu\", \"data\": \"" + command + "\"}"
}

// Used only for submenu
func replyWithMenu(item string) tgbotapi.InlineKeyboardMarkup {
	for key, value := range flatMenu {
		if key == "."+item {
			if value.Flatmenu != nil {
				return generateMenuKeyboard(value.Submenu, true)
			}
		}
	}
	return tgbotapi.NewInlineKeyboardMarkup()
}

func generateMenuKeyboard(menu []Menu, addBack bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, item := range menu {
		var buttonRow []tgbotapi.InlineKeyboardButton
		if item.Url != "" {
			buttonRow = append(buttonRow, tgbotapi.NewInlineKeyboardButtonURL(item.Title, item.Url))
		} else {
			buttonRow = append(buttonRow, tgbotapi.NewInlineKeyboardButtonData(item.Title, prepareCallbackMenuCommand(item.Command)))
		}
		keyboard = append(keyboard, buttonRow)
	}
	if addBack {
		var buttonRow []tgbotapi.InlineKeyboardButton
		buttonRow = append(buttonRow, tgbotapi.NewInlineKeyboardButtonData("↩️", "{\"command\": \"show_root_menu\"}"))
		keyboard = append(keyboard, buttonRow)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}
