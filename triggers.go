package main

import (
	"log"
	"os"
	"sort"
	"strings"

	tgbotapi "github.com/Alexkurd/telegram-bot-api/v7"
	"gopkg.in/yaml.v3"
)

type Trigger struct {
	Name       string      `yaml:"name"`
	Action     string      `yaml:"action"`
	ActionText string      `yaml:"actiontext"`
	Conditions []Condition `yaml:"conditions"`
}

type Condition struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type combotTrigger struct {
	Trigger []oldTrigger `yaml:"triggers"`
	Section []Section    `yaml:"sections"`
}

type oldTrigger struct {
	Name        string         `yaml:"name"`
	Conditions  []oldCondition `yaml:"condition"`
	Actions     string         `yaml:"actiontext"`
	ShowPreview bool           `default:"false" yaml:"showpreview"`
	Picture     string         `default:"" yaml:"picture"`
	Section     string         `yaml:"section"`
}

type oldCondition struct {
	Value string `yaml:"word"`
}

type Section struct {
	Id   string `yaml:"id"`
	Name string `yaml:"name"`
}

var oldconfig combotTrigger

func CheckTriggerMessage(message *tgbotapi.Message) bool {
	triggered := false
	for _, trigger := range oldconfig.Trigger {
		for _, condition := range trigger.Conditions {
			if strings.EqualFold(message.Text, condition.Value) {
				triggered = true
				continue
			}
		}
		if triggered {
			trigger.Actions = strings.Replace(trigger.Actions, "{reply_to_namelink}", getNameLink(*message.From), -1)
			//msg := tgbotapi.NewPhoto(message.Chat.ID, trigger.Picture)
			msg := tgbotapi.NewMessage(message.Chat.ID, trigger.Actions)
			//msg.Entities = append(msg.Entities, )
			msg.ReplyParameters.MessageID = message.MessageID
			if !trigger.ShowPreview {
				msg.LinkPreviewOptions.IsDisabled = true
			}
			msg.ParseMode = "HTML"
			if emulate {
				log.Print("Emulate:TriggeredGood:", message.Text)
				triggered = false
				continue
			}
			_, err := bot.Request(msg)
			if err != nil {
				log.Print("TriggeredBad: ", err, message.Text)
			}
			triggered = false
			log.Print("TriggeredGood:", message.Text)
		}
	}
	return triggered
}

func readTriggers() {
	configFile, err := os.ReadFile("triggers.yaml")
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(configFile, &oldconfig)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Triggers loaded: ", len(oldconfig.Trigger))
	log.Println("Sections loaded: ", len(oldconfig.Section))

	sort.Slice(oldconfig.Trigger, func(i, j int) bool {
		return oldconfig.Trigger[i].Name < oldconfig.Trigger[j].Name
	})
}

func getSectionsList() []Section {
	return oldconfig.Section
}

func getTriggersList() string {
	message := ""
	sections := getSectionsList()
	sectionTriggers := make(map[string]string)
	for _, trigger := range oldconfig.Trigger {
		if len(trigger.Conditions) > 0 {
			var words []string
			for _, condition := range trigger.Conditions {
				words = append(words, condition.Value)
			}
			//message = message + trigger.Name + ": " + strings.Join(words, "|") + "\r\n"
			if len(trigger.Section) > 0 {
				sectionTriggers[trigger.Section] = sectionTriggers[trigger.Section] + trigger.Name + ": " + strings.Join(words, "|") + "\r\n"
			}
		}
	}

	for _, section := range sections {
		if len(sectionTriggers[section.Id]) > 0 {
			message = message + "<b>" + section.Name + "</b>\r\n" + sectionTriggers[section.Id] + "\r\n"
		}
	}

	return message
}
