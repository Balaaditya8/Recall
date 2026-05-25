package main

import (
	"fmt"
	"log"
	"os"
	"recall/services"

	"recall/models"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	godotenv.Load(".env")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	db_url := os.Getenv("DB_URL")
	api := slack.New(botToken,
		slack.OptionAppLevelToken(appToken),
	)
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
					if result.Confidence == "high" {
						err = services.SaveDecision(db, result)
						if err != nil {
							fmt.Println("db error:", err)
							return
						}
					}
					//fmt.Printf("type: %s\nsummary: %s\nconfidence: %s\n", result.Type, result.Summary, result.Confidence)
				}()
			}
		}
	}()
	client.Run()
}
