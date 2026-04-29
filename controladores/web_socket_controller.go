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

	if nickname == "" || salaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname y sala_id requeridos"})
		return
	}

	// Validar que el usuario tenga acceso a la sala
	collection := db.GetCollection("usuarios")
	count, err := collection.CountDocuments(c.Request.Context(), bson.M{
		"sala_id":  salaID,
		"nickname": nickname,
		"activo":   true,
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

	// Validar que el usuario no esté ya conectado en otra sala (usando Redis)
	activeRoom, errRedis := db.RedisClient.Get(c.Request.Context(), "user_active_room:"+nickname).Result()
	if errRedis == nil && activeRoom != "" && activeRoom != salaID {
		conn.WriteJSON(sockets.Mensaje{
			Tipo:  "error",
			Texto: "Ya te encuentras conectado en la sala '" + activeRoom + "'. Usa un solo dispositivo o cierra esa sesión primero para ingresar a esta.",
		})
		conn.Close()
		return
	}

	cliente := &sockets.Cliente{
		Hub:      hub,
		Conn:     conn,
		Envio:    make(chan sockets.Mensaje, 256),
		Nickname: nickname,
		SalaId:   salaID,
		Ip:       c.ClientIP(),
	}

	hub.Registro <- cliente

	go cliente.WritePump()
	go cliente.ReadPump()
}
