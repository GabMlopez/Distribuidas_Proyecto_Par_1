package controladores

import (
	"chat_distribuido/controladores/sockets"
	"chat_distribuido/db"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetHistorialMensajes(c *gin.Context) {
	salaID := c.Param("roomId")
	if salaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId es requerido"})
		return
	}

	// Parámetros de paginación
	limit := int64(50)
	offset := int64(0)

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Obtener historial
	messages, err := db.ObtenerHistorialMensajes(salaID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener mensajes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    len(messages),
	})
}

func HandleWebSocket(hub *sockets.Hub, c *gin.Context) {
	nickname := c.Query("nickname")
	salaID := c.Query("sala_id")
	deviceID := c.Query("device_id")

	if nickname == "" || salaID == "" || deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname, sala_id y device_id requeridos"})
		return
	}

	collection := db.GetCollection("usuarios")

	// Buscar usuario existente con este device_id y sala_id
	var usuario struct {
		Nickname string `bson:"nickname"`
		DeviceID string `bson:"device_id"`
		SalaID   string `bson:"sala_id"`
		Activo   bool   `bson:"activo"`
		Ip       string `bson:"ip"`
	}

	err := collection.FindOne(c.Request.Context(), bson.M{
		"sala_id":   salaID,
		"device_id": deviceID,
	})

	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes acceso a esta sala o el nickname no está activo"})
		return
	}
	}).Decode(&usuario)

	clientIP := c.ClientIP()

	// Si el usuario existe y está activo pero es el mismo device_id, permitir reconexión
	if err == nil && usuario.Activo && usuario.DeviceID == deviceID {
		// Mismo dispositivo reconectando - permitir
		log.Printf("Reconectando dispositivo %s en sala %s", deviceID, salaID)

		// Limpiar sesión anterior en Redis para esta IP
		db.RedisClient.Del(c.Request.Context(), "device_active_session:"+usuario.Ip)

		// Actualizar a activo con nueva IP
		_, err = collection.UpdateOne(
			c.Request.Context(),
			bson.M{"sala_id": salaID, "device_id": deviceID},
			bson.M{
				"$set": bson.M{
					"activo":      true,
					"ip":          clientIP,
					"last_active": time.Now(),
				},
			},
		)
		if err != nil {
			log.Printf("Error actualizando usuario: %v", err)
		}
	} else if err == nil && usuario.Activo && usuario.DeviceID != deviceID {
		// Otro dispositivo con el mismo nickname - NO permitir
		c.JSON(http.StatusForbidden, gin.H{"error": "Este nickname ya está activo en otro dispositivo"})
		return
	} else if err != nil {
		// Usuario no existe - crearlo
		_, err = collection.InsertOne(c.Request.Context(), bson.M{
			"sala_id":     salaID,
			"nickname":    nickname,
			"device_id":   deviceID,
			"activo":      true,
			"ip":          clientIP,
			"joined_at":   time.Now(),
			"last_active": time.Now(),
		})
		if err != nil {
			log.Printf("Error creando usuario: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear usuario"})
			return
		}
	} else if err == nil && !usuario.Activo {
		// Usuario existe pero no está activo - reactivar
		_, err = collection.UpdateOne(
			c.Request.Context(),
			bson.M{"sala_id": salaID, "device_id": deviceID},
			bson.M{
				"$set": bson.M{
					"activo":      true,
					"ip":          clientIP,
					"last_active": time.Now(),
				},
			},
		)
		if err != nil {
			log.Printf("Error reactivando usuario: %v", err)
		}
	}

	// Verificar sesión activa en Redis para esta IP
	activeSession, errRedis := db.RedisClient.Get(c.Request.Context(), "device_active_session:"+clientIP).Result()
	if errRedis == nil && activeSession != "" && activeSession != deviceID {
		conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
		conn.WriteJSON(sockets.Mensaje{
			Tipo:  "error",
			Texto: "Este dispositivo ya tiene una sesión activa",
		})
		conn.Close()
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al establecer conexión WebSocket"})
		return
	}

	db.RedisClient.Set(c.Request.Context(), "device_active_session:"+clientIP, deviceID, 24*time.Hour)

	cliente := &sockets.Cliente{
		Hub:      hub,
		Conn:     conn,
		Envio:    make(chan sockets.Mensaje, 256),
		Nickname: nickname,
		SalaId:   salaID,
		DeviceId: deviceID,
		Ip:       clientIP,
	}

	hub.Registro <- cliente

	go cliente.WritePump()
	go cliente.ReadPump()
}
