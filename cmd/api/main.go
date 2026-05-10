package main

import (
	"chat_distribuido/internal/handlers"
	"chat_distribuido/internal/websocket"
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/middleware"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error cargando .env file")
	}

	repository.ConnectDB()
	repository.ConnectRedis()
	repository.ConnectMinio()
	defer repository.DisconnectDB()

	// Limpiar usuarios fantasma de sesiones anteriores (útil en desarrollo)
	collection := repository.GetCollection("usuarios")
	collection.UpdateMany(context.Background(), bson.M{}, bson.M{"$set": bson.M{"activo": false}})

	// Limpiar sesiones activas de Redis (stale keys de ejecuciones anteriores)
	iter := repository.RedisClient.Scan(context.Background(), 0, "device_active_session:*", 100).Iterator()
	for iter.Next(context.Background()) {
		repository.RedisClient.Del(context.Background(), iter.Val())
	}
	log.Println("Sesiones anteriores limpiadas (MongoDB + Redis)")

	// Eliminar índice antiguo basado en IP si existe
	_, _ = collection.Indexes().DropOne(context.Background(), "unique_active_ip")

	// Crear índice único parcial: solo puede existir 1 DeviceID activo a la vez
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "device_id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"activo": true}).
			SetName("unique_active_device_id"),
	}
	_, errIdx := collection.Indexes().CreateOne(context.Background(), indexModel)
	if errIdx != nil {
		log.Printf("Aviso índice único DeviceID: %v (puede que ya exista)", errIdx)
	} else {
		log.Println("Índice único parcial (device_id + activo:true) creado/verificado")
	}

	// Crear hub central de WebSocket
	hub := websocket.Nuevo_Hub()
	go hub.Run()
	handlers.SetHub(hub)

	// Configurar router
	r := gin.Default()
	r.SetTrustedProxies(nil) // Fix: No confiar en todos los proxies por defecto (Seguridad)

	// Middleware para bloquear localhost (IP y Origin) excepto /health
	r.Use(func(c *gin.Context) {
		// Permitir healthchecks de Render/Docker internamente
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		origin := c.Request.Header.Get("Origin")

		if ip == "127.0.0.1" || ip == "::1" || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Acceso denegado: Usa la IP de la red local (ej. 192.168.x.x) en lugar de localhost.",
			})
			c.Abort()
			return
		}
		c.Next()
	})

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
		// Permitir específicamente tu frontend en Vercel
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "https://distribuidas-proyecto-par-1.vercel.app/")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Rutas públicas
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Autenticación
	auth := r.Group("/auth")
	{
		auth.POST("/login", handlers.Login_handler)
		auth.POST("/logout", handlers.Logout)
	}

	// Gestión de salas
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddlewareAdmin())
	{
		// Crear sala (solo admin)
		adminRoutes.POST("/rooms", handlers.CreateSalasHandler)

		// Actualizar sala (solo admin)
		adminRoutes.PUT("/rooms/:roomId", handlers.UpdateSalaHandler)

		// Eliminar sala (solo admin)
		adminRoutes.DELETE("/rooms/:roomId", handlers.DeleteSalaHandler)

		// Obtener todas las salas (admin)
		adminRoutes.GET("/rooms", handlers.GetAllSalasAdmin)
	}

	roomRoutes := r.Group("/rooms")
	{
		roomRoutes.GET("/list", handlers.ListaSalas)
		roomRoutes.POST("/join", handlers.UnirseSalaHandler)
		roomRoutes.POST("/leave", handlers.DejarSalaHandler)
		roomRoutes.GET("/:roomId/messages", handlers.GetMessagesHandler)
	}

	// WebSocket
	r.GET("/ws/:roomId", func(c *gin.Context) {
		handlers.HandleWebSocket(hub, c)
	})

	// Subida de archivos (Protegido para subir, público para descargar vía link)
	r.GET("/upload/file/:filename", handlers.GetFileHandler)
	uploadRoutes := r.Group("/upload")
	uploadRoutes.Use(middleware.AuthMiddlewareUser())
	{
		uploadRoutes.POST("/file", handlers.UploadFileHandler)
		uploadRoutes.DELETE("/file/:filename", handlers.DeleteFileHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor corriendo en el puerto %s", port)
	r.Run(":" + port)
}
