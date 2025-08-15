package main

import (
	"github.com/joho/godotenv"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, continuing...")
	}
}

func main() {

}
