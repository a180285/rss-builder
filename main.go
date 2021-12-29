package main

import (
	"github.com/a180285/rss-builder/src/ecnu"
	"github.com/a180285/rss-builder/src/gin_middlewares"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(gin_middlewares.MyGinRecovery)

	router.GET("/ecnu-contest/rss", ecnu.BuildFeeds)

	router.Run()
}
