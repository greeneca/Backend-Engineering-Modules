package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)


func main() {
	consumeWikipediaChanges()
}

func consumeWikipediaChanges() {
	rsp, err := http.Get("https://stream.wikimedia.org/v2/stream/recentchange")
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	for {
		line := make([]byte, 1024) // Buffer to read the stream
		_, err := rsp.Body.Read(line)
		if err != nil {
			panic(err)
		}
		if len(line) > 0 {
			fmt.Println(string(line))
		}
	}
}

func server() {
	router := gin.Default()
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Run(":7000")
}
