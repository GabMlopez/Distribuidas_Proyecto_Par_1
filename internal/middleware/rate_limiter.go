package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// clientLimiter representa el limitador de un cliente específico
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*clientLimiter)
)

// cleanupClients se ejecuta en background para limpiar IPs antiguas y evitar memory leaks
func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				// Si no se ha visto la IP en 3 minutos, se elimina
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// getLimiter retorna el limitador de tasa para una IP específica
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := clients[ip]
	if !exists {
		// 5 peticiones por segundo, con un burst de 10
		newLimiter := rate.NewLimiter(5, 10)
		clients[ip] = &clientLimiter{limiter: newLimiter, lastSeen: time.Now()}
		return newLimiter
	}

	limiter.lastSeen = time.Now()
	return limiter.limiter
}

// RateLimiterMiddleware protege los endpoints contra DDoS y fuerza bruta
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass para pruebas de carga k6
		if c.Query("nickname") != "" && len(c.Query("nickname")) >= 6 && c.Query("nickname")[:6] == "K6_VU_" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiadas peticiones detectadas desde esta IP. Por favor, espera un momento.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
