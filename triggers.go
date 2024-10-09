package main

import (
	"log"
	"os"
	"regexp"
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
	Name           string         `yaml:"name"`
	Conditions     []oldCondition `yaml:"condition"`
	CheckSubstring bool           `default:"false" yaml:"substringSearch"`
	CheckRegexp    bool           `default:"false" yaml:"regexpSearch"`
	Actions        string         `yaml:"actiontext"`
	ShowPreview    bool           `default:"false" yaml:"showpreview"`
	Picture        string         `default:"" yaml:"picture"`
	Section        string         `yaml:"section"`
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
			if isQuestion(message.Text) && trigger.CheckSubstring {
				if strings.Contains(strings.ToLower(message.Text), strings.ToLower(condition.Value)) {
					triggered = true
					continue
				}
			}
			if isQuestion(message.Text) && trigger.CheckRegexp {
				regex := regexp.MustCompile(condition.Value)
				if regex.MatchString(message.Text) {
					triggered = true
					continue
				}
			}
		}
		if triggered {
			trigger.Actions = strings.Replace(trigger.Actions, "{reply_to_namelink}", getNameLink(*message.From), -1)
			var msg tgbotapi.Chattable
			if len(trigger.Picture) > 0 {
				photoConfig := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileID(trigger.Picture))
				photoConfig.Caption = trigger.Actions
				photoConfig.ParseMode = "HTML"
				msg = photoConfig
			} else {
				messageConfig := tgbotapi.NewMessage(message.Chat.ID, trigger.Actions)
				messageConfig.ParseMode = "HTML"
				messageConfig.ReplyParameters.MessageID = message.MessageID
				if !trigger.ShowPreview {
					messageConfig.LinkPreviewOptions.IsDisabled = true
				}
				msg = messageConfig
			}

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

func isQuestion(message string) bool {
	return strings.Contains(message, "?")
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
