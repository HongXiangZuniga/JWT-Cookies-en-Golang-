package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
)

func main() {
	router.POST("/login", Login)
	router.POST("/todo", CreateTodo)
	router.POST("/logout", Logout)

	log.Fatal(router.Run(":8080"))
}
