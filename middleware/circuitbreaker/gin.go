// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinMiddleware is a middleware for circuit breaker in Gin framework.
func GinMiddleware(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.FullPath()
		breaker := mgr.Get(key)

		_, err := breaker.Execute(func() (any, error) {
			c.Next()
			if len(c.Errors) > 0 {
				return nil, c.Errors.Last()
			}
			return nil, nil
		})

		if err != nil {
			c.Errors = append(c.Errors, &gin.Error{Err: err})
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, nil)
			return
		}
	}
}
