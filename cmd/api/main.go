package main

import (
	"github.com/gin-gonic/gin"
	gdb "gn-indexer/internal/db"
)

func main() {
	db := gdb.MustConnect()
	_ = db //repo에 주입에서 사용

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.Run(":8080")
}
