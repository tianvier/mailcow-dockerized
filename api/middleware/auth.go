package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const TOKEN = "4274B4B8329E47705C11B9E1A9967170"

func AuthToken() gin.HandlerFunc {

	authToken := TOKEN
	if os.Getenv("API_TOKEN") != "" {
		authToken = os.Getenv("API_TOKEN")
	}

	return func(c *gin.Context) {
		token := c.Request.Header.Get("x-token")
		//_, err := models.GetUserById(userId)
		if token != authToken {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "未登录或非法访问",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
