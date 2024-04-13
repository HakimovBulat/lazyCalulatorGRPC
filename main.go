package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	router := gin.Default()
	router.GET("/", mainHandler)
	router.LoadHTMLGlob("templates/*.html")
	if err := router.Run(":8080"); err != nil {
		logger.Error(err.Error(), zap.String("router", "f*ck down"))
	}
}
func mainHandler(c *gin.Context) {
	cookie, err := c.Request.Cookie("user")
	fmt.Println(cookie, err)
	if err != nil {
		cookie = &http.Cookie{
			Name:   "user",
			Value:  "12345",
			MaxAge: 300,
		}
		http.SetCookie(c.Writer, cookie)
	}
	c.HTML(200, "index.html", cookie)
}
