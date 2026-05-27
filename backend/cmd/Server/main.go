package main

import (
	"database/sql"
	"log"
	"os"
	"recall/handlers"
	"recall/models"
	"recall/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	handlers.InitSlack(botToken, appToken)
	db_url := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/slack/events", handlers.HandleSlackEvents)
	r.GET("/decisions", func(c *gin.Context) {
		decisions, err := services.GetDecisions(db)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, decisions)
	})

	r.POST("/decisions/:id/confirm", func(c *gin.Context) {
		id := c.Param("id")
		var event models.ExtractedEvent
		c.BindJSON(&event)
		err := services.ConfirmDecision(db, id, event)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "confirmed"})
	})

	r.POST("/decisions/:id/dismiss", func(c *gin.Context) {
		id := c.Param("id")
		err := services.DismissDecision(db, id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "dismissed"})
	})

	r.Run(":8080")
}
