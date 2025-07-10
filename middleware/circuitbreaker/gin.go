// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinCircuitBreakerHandler is a middleware for circuit breaker in Gin framework.
func GinCircuitBreakerHandler(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.FullPath()
		breaker := mgr.Get(key)

		_, err := breaker.Execute(func() (interface{}, error) {
			c.Next()
			if len(c.Errors) > 0 {
				return nil, c.Errors.Last()
			}
			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "circuit breaker triggered: " + err.Error(),
			})
			return
		}
	}
}
