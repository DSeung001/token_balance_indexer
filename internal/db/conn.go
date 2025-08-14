package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func MustConnect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
