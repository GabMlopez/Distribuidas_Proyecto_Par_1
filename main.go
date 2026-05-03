package main

import (
	"chat_distribuido/controladores"
	"chat_distribuido/controladores/sockets"
	"chat_distribuido/db"
	"chat_distribuido/middleware"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error cargando .env file")
	}

	db.ConnectDB()
	db.ConnectRedis()
	db.ConnectMinio()
	defer db.DisconnectDB()

	// Limpiar usuarios fantasma de sesiones anteriores (útil en desarrollo)
	collection := db.GetCollection("usuarios")
	collection.UpdateMany(context.Background(), bson.M{}, bson.M{"$set": bson.M{"activo": false}})

	// Crear hub central de WebSocket
	hub := sockets.Nuevo_Hub()
	go hub.Run()
	controladores.SetHub(hub)

	// Configurar router
	r := gin.Default()
	r.SetTrustedProxies(nil) // Fix: No confiar en todos los proxies por defecto (Seguridad)

	// Configurar CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Configurar Security Headers (CSP, X-Frame-Options, X-Content-Type-Options)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Next()
	})

	// Rutas públicas
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Autenticación
	auth := r.Group("/auth")
	{
		auth.POST("/login", controladores.Login_handler)
		auth.POST("/logout", controladores.Logout)
	}

	// Gestión de salas
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddlewareAdmin())
	{
		// Crear sala (solo admin)
		adminRoutes.POST("/rooms", controladores.CreateSalasHandler)

		// Actualizar sala (solo admin)
		adminRoutes.PUT("/rooms/:roomId", controladores.UpdateSalaHandler)

		// Eliminar sala (solo admin)
		adminRoutes.DELETE("/rooms/:roomId", controladores.DeleteSalaHandler)

		// Obtener todas las salas (admin)
		adminRoutes.GET("/rooms", controladores.GetAllSalasAdmin)
	}

	roomRoutes := r.Group("/rooms")
	{
		roomRoutes.GET("/list", controladores.ListaSalas)
		roomRoutes.POST("/join", controladores.UnirseSalaHandler)
		roomRoutes.POST("/leave", controladores.DejarSalaHandler)
	}

	// WebSocket
	r.GET("/ws/:roomId", func(c *gin.Context) {
		controladores.HandleWebSocket(hub, c)
	})

	// Subida de archivos (Protegido para subir, público para descargar vía link)
	r.GET("/upload/file/:filename", controladores.GetFileHandler)
	uploadRoutes := r.Group("/upload")
	uploadRoutes.Use(middleware.AuthMiddlewareUser())
	{
		uploadRoutes.POST("/file", controladores.UploadFileHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor corriendo en el puerto %s", port)
	r.Run(":" + port)
}
