package middleware

import (
	"chat_distribuido/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddlewareAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token no proporcionado",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// Verificar formato Bearer
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Formato de token inválido. Use: Bearer <token>",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// Validar el token usando la función unificada
		token, claims, err := utils.VerificarTokenAdmin(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido: " + err.Error(),
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Verificar si el token es válido
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expirado o inválido",
				"code":  "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// Guardar claims en el contexto
		c.Set("admin_id", claims["admin_id"])
		c.Set("usuario", claims["usuario"])
		c.Set("exp", claims["exp"])

		c.Next()
	}
}

func AuthMiddlewareUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token no proporcionado",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// Verificar formato Bearer
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Formato de token inválido. Use: Bearer <token>",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// Validar el token usando la función unificada
		token, claims, err := utils.VerificarTokenUser(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido: " + err.Error(),
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Verificar si el token es válido
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expirado o inválido",
				"code":  "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// Guardar claims en el contexto
		c.Set("usuario", claims["usuario"])
		c.Set("exp", claims["exp"])

		c.Next()
	}
}
