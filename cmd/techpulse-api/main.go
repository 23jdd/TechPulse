package main

import (
	"config"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	c := config.MustLoadConfig()
	if err := router.Run(fmt.Sprintf("%s:%d", c.ApiAddress, c.ApiPort)); err != nil {
		panic(err)
	}
}
