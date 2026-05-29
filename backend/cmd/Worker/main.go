package main

import (
	"fmt"
	"log"
	"os"
	"recall/handlers"
	"recall/services"
	"strings"

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

				isReply := innerEvent.ThreadTimeStamp != ""
				wordCount := len(strings.Fields(innerEvent.Text))

				if !isReply && wordCount <= 5 {
					// skip, not worth processing
					continue // or return
				}

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
					params := slack.GetConversationHistoryParameters{
						ChannelID: innerEvent.Channel,
						Limit:     5,
					}
					history, _ := handlers.Client.GetConversationHistory(&params)
					msgs := history.Messages
					for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
						msgs[i], msgs[j] = msgs[j], msgs[i]
					}
					for _, m := range msgs[:len(msgs)-1] {
						contextMessages = append(contextMessages, m.Text)
					}
				}

				message := models.SlackMessage{Text: innerEvent.Text,
					Channel:   innerEvent.Channel,
					Timestamp: innerEvent.TimeStamp,
					User:      innerEvent.User,
				}

				go func() {
					existing, err := services.GetRecentDecisions(db, 10)
					if err != nil {
						fmt.Println("error fetching existing decisions:", err)
						existing = []models.ExtractedEvent{} // empty slice, don't block
					}
					result, err := services.ProcessWithOllama(message, contextMessages, existing)
					if err != nil {
						fmt.Println("ollama error:", err)
						return
					}
					fmt.Print(result)
					if result.Status == "update" {
						services.UpdateDecision(db, result)
					} else if result.Type != "" && result.Status == "confirmed" {
						services.SaveDecision(db, result)
					} else if result.Type != "" && result.Status == "pending" {
						services.SaveDecision(db, result)
					}
					//fmt.Printf("type: %s\nsummary: %s\nconfidence: %s\n", result.Type, result.Summary, result.Confidence)
				}()
			}
		}
	}()
	client.Run()
}
