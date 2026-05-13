package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Limitador de tasa por IP
var (
	ips = make(map[string]*rate.Limiter)
	mu  sync.Mutex
)

// GetLimiter obtiene o crea un limitador para una IP
// Limitaremos a 5 requests por segundo con un burst (ráfaga) de 10
func GetLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := ips[ip]
	if !exists {
		// 5 requests per second, burst of 10
		limiter = rate.NewLimiter(5, 10)
		ips[ip] = limiter
	}

	return limiter
}

// RateLimiter es el middleware para prevenir DDoS y Fuerza Bruta
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiadas peticiones (Estás siendo limitado temporalmente - Prevención DDoS)",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
