package main

import (
	"github.com/gin-gonic/gin"
	"github.com/linweiyuan/go-chatgpt-api/api"
	"github.com/linweiyuan/go-chatgpt-api/api/chatgpt"
	"github.com/linweiyuan/go-chatgpt-api/api/platform"
	_ "github.com/linweiyuan/go-chatgpt-api/env"
	"github.com/linweiyuan/go-chatgpt-api/middleware"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strings"
)

func init() {
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	//启动时获取一次token
	chatgpt.InitToken()
	middleware.Init()
	//定时任务,每天凌晨1点更新token
	crontab := cron.New(cron.WithSeconds())
	// 添加定时任务,
	crontab.AddFunc("0 0 1 * * ? ", chatgpt.InitToken)
	// 更新第三方调用授权码
	crontab.AddFunc("0 0/1 * * * ? *", middleware.Init)
	// 启动定时器
	crontab.Start()
}

//goland:noinspection SpellCheckingInspection
func main() {

	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.CheckHeaderMiddleware())

	setupChatGPTAPIs(router)
	setupPlatformAPIs(router)
	setupPandoraAPIs(router)
	router.NoRoute(api.Proxy)

	router.GET("/healthCheck", api.HealthCheck)

	port := os.Getenv("GO_CHATGPT_API_PORT")
	if port == "" {
		port = "8080"
	}
	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server: " + err.Error())
	}
}

//goland:noinspection SpellCheckingInspection
func setupChatGPTAPIs(router *gin.Engine) {
	chatgptGroup := router.Group("/chatgpt")
	chatgptGroup.POST("/login", chatgpt.LoginApi)
	chatgptGroup.POST("/backend-api/conversation", chatgpt.CreateConversation)
	chatgptGroup.POST("/backend-api/conversation/simple", chatgpt.CreateConversationSimple)
}

func setupPlatformAPIs(router *gin.Engine) {
	platformGroup := router.Group("/platform")
	{
		platformGroup.POST("/login", chatgpt.LoginApi)
		apiGroup := platformGroup.Group("/v1")
		apiGroup.POST("/chat/completions", platform.CreateChatCompletions)
		apiGroup.POST("/completions", platform.CreateCompletions)
	}
}

//goland:noinspection SpellCheckingInspection
func setupPandoraAPIs(router *gin.Engine) {
	pandoraEnabled := os.Getenv("GO_CHATGPT_API_PANDORA") != ""
	if pandoraEnabled {
		router.GET("/api/*path", func(c *gin.Context) {
			c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, "/api", "/chatgpt/backend-api")
			router.HandleContext(c)
		})
		router.POST("/api/*path", func(c *gin.Context) {
			c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, "/api", "/chatgpt/backend-api")
			router.HandleContext(c)
		})
	}
}
