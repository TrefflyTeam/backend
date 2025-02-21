package main

import _ "github.com/jackc/pgx/v5"
import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
