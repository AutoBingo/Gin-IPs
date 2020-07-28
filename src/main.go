package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	route := gin.Default()
	route_request.InitValidator()
	route.GET("/", v1_sdk.SearchIpHandlerWithGet)
	if err := route.Run("127.0.0.1:8080"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
