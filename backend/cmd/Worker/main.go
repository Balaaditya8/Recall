package main

import (
	"fmt"
	"os"
	"recall/services"

	"recall/models"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	godotenv.Load(".env")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	api := slack.New(botToken,
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(api)
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				client.Ack(*evt.Request)
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					continue
				}

				innerEvent, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.MessageEvent)
				if !ok {
					continue
				}

				message := models.SlackMessage{Text: innerEvent.Text,
					Channel:   innerEvent.Channel,
					Timestamp: innerEvent.TimeStamp,
					User:      innerEvent.User,
				}

				go func() {
					result, err := services.ProcessWithOllama(message)
					if err != nil {
						fmt.Println("ollama error:", err)
						return
					}
					fmt.Print(result)
					//fmt.Printf("type: %s\nsummary: %s\nconfidence: %s\n", result.Type, result.Summary, result.Confidence)
				}()
			}
		}
	}()
	client.Run()
}
