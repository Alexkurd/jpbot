package main

import (
	"log"
	"os"
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
}

type oldTrigger struct {
	Name        string         `yaml:"name"`
	Conditions  []oldCondition `yaml:"condition"`
	Actions     string         `yaml:"actiontext"`
	ShowPreview bool           `default:"false" yaml:"showpreview"`
}

type oldCondition struct {
	Value string `yaml:"word"`
}

var oldconfig combotTrigger

func CheckTriggerMessage(message *tgbotapi.Message) {
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
			msg := tgbotapi.NewMessage(message.Chat.ID, trigger.Actions)
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
}

func readTriggers() {
	configFile, err := os.ReadFile("triggers.yaml")
	if err != nil {
		log.Panic(err)
	}
	//var config Config
	err = yaml.Unmarshal(configFile, &oldconfig)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Triggers loaded: ", len(oldconfig.Trigger))
}
