package middleware

import (
	"github.com/gothms/httpgo/framework/gin"
	"log"
	"time"
)

// Cost recover 机制，将协程中的函数异常进行捕获
func Cost() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		log.Printf("api uri start: %v", c.Request.RequestURI)
		c.Next()
		cost := time.Now().Sub(start)
		log.Printf("api uri: %v,cost: %v", c.Request.RequestURI, cost.Seconds())
	}
}
