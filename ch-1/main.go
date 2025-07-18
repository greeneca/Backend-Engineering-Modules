package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func main() {
	server()
}

func server() {
	router := gin.Default()
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Run(":7000")
}
