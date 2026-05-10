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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error cargando .env file")
	}

	db.ConnectDB()
	db.ConnectRedis()
	db.ConnectMinio()
	db.CrearIndicesMensajes()
	defer db.DisconnectDB()

	// Limpiar usuarios fantasma de sesiones anteriores (útil en desarrollo)
	collection := db.GetCollection("usuarios")
	collection.UpdateMany(context.Background(), bson.M{}, bson.M{"$set": bson.M{"activo": false}})

	// Limpiar sesiones activas de Redis (stale keys de ejecuciones anteriores)
	iter := db.RedisClient.Scan(context.Background(), 0, "device_active_session:*", 100).Iterator()
	for iter.Next(context.Background()) {
		db.RedisClient.Del(context.Background(), iter.Val())
	}
	log.Println("Sesiones anteriores limpiadas (MongoDB + Redis)")

	// Crear índice único parcial: solo puede existir 1 IP activa a la vez
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "ip", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"activo": true}).
			SetName("unique_active_ip"),
	}
	_, errIdx := collection.Indexes().CreateOne(context.Background(), indexModel)
	if errIdx != nil {
		log.Printf("Aviso índice único IP: %v (puede que ya exista)", errIdx)
	} else {
		log.Println("Índice único parcial (ip + activo:true) creado/verificado")
	}

	// Crear hub central de WebSocket
	hub := sockets.Nuevo_Hub()
	go hub.Run()
	controladores.SetHub(hub)

	// Configurar router
	r := gin.Default()
	r.SetTrustedProxies(nil) // Fix: No confiar en todos los proxies por defecto (Seguridad)

	// Configurar CORS
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Permitir orígenes específicos
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
		}

		// Verificar si el origen está permitido
		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		// Para desarrollo, permitir cualquier localhost
		if !isAllowed && (len(origin) == 0 || origin == "null") {
			isAllowed = true
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, X-Requested-With, Accept, Connection, Upgrade")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// Cabeceras específicas para WebSocket
		if c.Request.Header.Get("Upgrade") == "websocket" {
			c.Writer.Header().Set("Connection", "Upgrade")
			c.Writer.Header().Set("Upgrade", "websocket")
		}

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
		roomRoutes.GET("/:roomId/messages", controladores.GetHistorialMensajes)

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
