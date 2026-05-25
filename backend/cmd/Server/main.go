package main

import (
	"database/sql"
	"log"
	"os"
	"recall/handlers"
	"recall/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")

	db_url := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := gin.Default()
	r.POST("/slack/events", handlers.HandleSlackEvents)
	r.GET("/decisions", func(c *gin.Context) {
		decisions, err := services.GetDecisions(db)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, decisions)
	})
	r.Run(":8080")
}
