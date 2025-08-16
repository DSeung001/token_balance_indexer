package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gn-indexer/internal/config"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {
	connConfig := config.NewDatabaseConfig()
	gormDb, err := connConfig.Connect()
	if err != nil {
		log.Fatal(err)
	}
	_ = gormDb

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.Run(":8080")
}
