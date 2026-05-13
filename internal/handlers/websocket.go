package handlers

import (
	"chat_distribuido/internal/repository"
	internalWs "chat_distribuido/internal/websocket"
	"log"
	"net/http"
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

func HandleWebSocket(hub *internalWs.Hub, c *gin.Context) {
	nickname := c.Query("nickname")
	salaID := c.Param("roomId") // Obtener param de la URL en vez de query
	if salaID == "" {
		salaID = c.Query("sala_id") // Fallback
	}
	deviceID := c.Query("device_id")

	if nickname == "" || salaID == "" || deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname, roomId y device_id requeridos"})
	}

	// Validar que el usuario tenga acceso a la sala
	collection := repository.GetCollection("usuarios")
	count, err := collection.CountDocuments(c.Request.Context(), bson.M{
		"sala_id":   salaID,
		"device_id": deviceID,
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

	clientIP := getRealIP(c) // Usar getRealIP para normalizar ::1 → 127.0.0.1

	// === CAPA DEFINITIVA: Triple validación (DeviceID + Nickname + IP) ===
	// Bloquea: misma pestaña, incógnito, otro navegador en la misma máquina.
	if hub.IsSessionBlocked(deviceID, nickname, clientIP) {
		conn.WriteJSON(internalWs.Mensaje{
			Tipo:  "error",
			Texto: "⚠️ Conexión Rechazada: Ya existe una sesión activa desde este dispositivo. No se permiten múltiples ventanas, pestañas, modo incógnito ni otros navegadores simultáneamente.",
		})
		conn.Close()
		return
	}

	// Validar que el dispositivo no tenga otra sesión activa (usando Redis)
	activeSession, errRedis := repository.RedisClient.Get(c.Request.Context(), "device_active_session:"+deviceID).Result()
	if errRedis == nil && activeSession != "" {
		expectedSession := nickname + "|" + salaID
		if activeSession != expectedSession {
			conn.WriteJSON(internalWs.Mensaje{
				Tipo:  "error",
				Texto: "⚠️ Conflicto de Estado: Redis detectó que tu dispositivo ya está anclado a una sesión distinta. Cierra la pestaña anterior o presiona el botón 'Salir de la sala' antes de reconectarte.",
			})
			conn.Close()
			return
		}
	}
	cliente := &internalWs.Cliente{
		Hub:      hub,
		Conn:     conn,
		Envio:    make(chan internalWs.Mensaje, 256),
		Nickname: nickname,
		SalaId:   salaID,
		DeviceId: deviceID,
		Ip:       clientIP,
	}

	hub.Registro <- cliente

	go cliente.WritePump()
	go cliente.ReadPump()
}
