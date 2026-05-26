package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

type SlackEvent struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	User    string `json:"user"`
	Channel string `json:"channel"`
}

func HandleSlackEvents(c *gin.Context) {
	var event SlackEvent
	err := c.BindJSON(&event)
	if err != nil {
		c.JSON(400, gin.H{"error": "could not read body"})
		return
	}
	fmt.Println(event.Text)

	c.JSON(200, gin.H{
		"status": "received slack event",
	})
}

var Client *slack.Client

func InitSlack(botToken, appToken string) {
	Client = slack.New(botToken, slack.OptionAppLevelToken(appToken))
}

var channelCache = map[string]string{}

func GetChannelName(channelID string) string {
	if name, exists := channelCache[channelID]; exists {
		return name
	}
	channel, err := Client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return channelID // fallback to ID if API fails
	}
	channelCache[channelID] = channel.Name
	return channel.Name
}
