package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func MustConnect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is empty")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
