package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
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
