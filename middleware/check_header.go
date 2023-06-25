package middleware

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"os"
)

var API_KEYS map[string]bool

func Init() {
	println("Init API Keys")
	API_KEYS = make(map[string]bool)
	if _, err := os.Stat("api_keys.txt"); err == nil {
		file, _ := os.Open("api_keys.txt")
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			key := scanner.Text()
			if key != "" {
				API_KEYS[key] = true
			}
		}
	}
}

//goland:noinspection SpellCheckingInspection
func CheckHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if len(API_KEYS) == 0 {
			c.Next()
			return
		}

		if API_KEYS[c.Request.Header.Get("Authorization")] {
			c.Next()
			return
		}
		c.String(401, "Unauthorized")
		c.Abort()
	}
}
