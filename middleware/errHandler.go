package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var Err error
				if e, ok := err.(error); ok {
					Err = e
				}
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": Err.Error(),
					"show":    true,
				})
				return
			}
		}()
		c.Next()
	}
}
