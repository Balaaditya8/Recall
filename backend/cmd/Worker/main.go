package main

import (
	"fmt"
	"log"
	"os"
	"recall/handlers"
	"recall/services"

	"recall/models"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var channelCache = map[string]string{}

func main() {
	godotenv.Load(".env")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	db_url := os.Getenv("DB_URL")
	handlers.InitSlack(botToken, appToken)
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	fmt.Println("connected to postgres")

	client := socketmode.New(handlers.Client)
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

				channel_name := handlers.GetChannelName(innerEvent.Channel)

				fmt.Println("Channel:", channel_name)

				var contextMessages []string

				if innerEvent.ThreadTimeStamp != "" {
					// it's a reply, fetch the whole thread
					params := slack.GetConversationRepliesParameters{
						ChannelID: innerEvent.Channel,
						Timestamp: innerEvent.ThreadTimeStamp,
					}
					msgs, _, _, err := handlers.Client.GetConversationReplies(&params)
					if err != nil {
						fmt.Println("error fetching thread:", err)
					} else {
						for _, m := range msgs[:len(msgs)-1] {
							contextMessages = append(contextMessages, m.Text)
						}
					}
				} else {
					// single message, no thread
					contextMessages = append(contextMessages, "")
				}

				message := models.SlackMessage{Text: innerEvent.Text,
					Channel:   innerEvent.Channel,
					Timestamp: innerEvent.TimeStamp,
					User:      innerEvent.User,
				}

				go func() {
					result, err := services.ProcessWithOllama(message, contextMessages)
					if err != nil {
						fmt.Println("ollama error:", err)
						return
					}
					fmt.Print(result)
					if result.Type == "none" {
						// ignore completely
					} else {
						services.SaveDecision(db, result)
					}
					//fmt.Printf("type: %s\nsummary: %s\nconfidence: %s\n", result.Type, result.Summary, result.Confidence)
				}()
			}
		}
	}()
	client.Run()
}
