package controladores

import (
	"chat_distribuido/controladores/sockets"
	"chat_distribuido/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(hub *sockets.Hub, c *gin.Context) {
	nickname := c.Query("nickname")
	salaID := c.Query("sala_id")
	deviceID := c.Query("device_id")

	if nickname == "" || salaID == "" || deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname, sala_id y device_id requeridos"})
		return
	}

	// Validar que el usuario tenga acceso a la sala
	collection := db.GetCollection("usuarios")
	count, err := collection.CountDocuments(c.Request.Context(), bson.M{
		"sala_id":   salaID,
		"nickname":  nickname,
		"device_id": deviceID,
	})

	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes acceso a esta sala o el nickname no está activo"})
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
		conn.WriteJSON(sockets.Mensaje{
			Tipo:  "error",
			Texto: "⚠️ Conexión Rechazada: Ya existe una sesión activa desde este dispositivo. No se permiten múltiples ventanas, pestañas, modo incógnito ni otros navegadores simultáneamente.",
		})
		conn.Close()
		return
	}

	// Validar que el dispositivo no tenga otra sesión activa (usando Redis)
	activeSession, errRedis := db.RedisClient.Get(c.Request.Context(), "device_active_session:"+deviceID).Result()
	if errRedis == nil && activeSession != "" {
		expectedSession := nickname + "|" + salaID
		if activeSession != expectedSession {
			conn.WriteJSON(sockets.Mensaje{
				Tipo:  "error",
				Texto: "⚠️ Conflicto de Estado: Redis detectó que tu dispositivo ya está anclado a una sesión distinta. Cierra la pestaña anterior o presiona el botón 'Salir de la sala' antes de reconectarte.",
			})
			conn.Close()
			return
		}
	}
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
