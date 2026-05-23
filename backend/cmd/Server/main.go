package main

import (
	"recall/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/slack/events", handlers.HandleSlackEvents)
	r.Run(":8080")
}
