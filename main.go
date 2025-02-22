package main

import _ "github.com/jackc/pgx/v5"
import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"hello": "world"}) })
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
