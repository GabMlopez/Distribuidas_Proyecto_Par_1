package sockets

import (
	"chat_distribuido/db"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Hub struct {
	Salas map[string]map[*Cliente]bool

	Registro chan *Cliente

	Desregistro chan *Cliente

	Broadcast chan Mensaje

	mutex       sync.RWMutex
	RedisClient *redis.Client
	ctx         context.Context
}

func Nuevo_Hub() *Hub {
	return &Hub{
		Salas:       make(map[string]map[*Cliente]bool),
		Registro:    make(chan *Cliente),
		Desregistro: make(chan *Cliente),
		Broadcast:   make(chan Mensaje),
		RedisClient: db.GetRedisClient(),
		ctx:         context.Background(),
	}
}

func (h *Hub) Run() {
	// Goroutine para suscribirse a Redis
	go h.listenRedis()

	for {
		select {
		case client := <-h.Registro:
			h.mutex.Lock()
			if _, ok := h.Salas[client.SalaId]; !ok {
				h.Salas[client.SalaId] = make(map[*Cliente]bool)
			}
			h.Salas[client.SalaId][client] = true
			h.mutex.Unlock()

			// Registrar en Redis la sala activa del usuario con una caducidad por seguridad (ej. 24h)
			h.RedisClient.Set(h.ctx, "user_active_room:"+client.Nickname, client.SalaId, 24*time.Hour)

			h.notifyUserList(client.SalaId)
			// Notificar que alguien se unió
			go func(c *Cliente) {
				h.Broadcast <- Mensaje{
					Tipo:     "join",
					Nickname: "Sistema",
					Texto:    c.Nickname + " se ha unido a la sala",
					SalaID:   c.SalaId,
				}
			}(client)
			log.Printf("Cliente %s conectado a sala %s", client.Nickname, client.SalaId)

			// Enviar historial de mensajes de la sala al nuevo cliente
			go h.cargarHistorialSala(client)

		case client := <-h.Desregistro:
			h.mutex.Lock()

			// Determinar si este usuario tiene otras conexiones en esta u otras salas
			userTieneMasConexiones := false

			if clients, ok := h.Salas[client.SalaId]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Envio)

					// Verificar si el usuario aún tiene otra conexión en esta sala
					for c := range clients {
						if c.Nickname == client.Nickname {
							userTieneMasConexiones = true
							break
						}
					}

					// Si la sala queda vacía, la eliminamos
					if len(clients) == 0 {
						delete(h.Salas, client.SalaId)
					}
				}
			}
			h.mutex.Unlock()

			// Eliminar registro de sala en Redis solo si ya no tiene conexiones activas
			if !userTieneMasConexiones {
				h.RedisClient.Del(h.ctx, "user_active_room:"+client.Nickname)
			}

			// Notificar usuarios actualizados
			h.notifyUserList(client.SalaId)
			// Notificar que alguien salió
			go func(c *Cliente) {
				h.Broadcast <- Mensaje{
					Tipo:     "leave",
					Nickname: "Sistema",
					Texto:    c.Nickname + " ha salido de la sala",
					SalaID:   c.SalaId,
				}
			}(client)
			log.Printf("Cliente %s desconectado de sala %s", client.Nickname, client.SalaId)

		case message := <-h.Broadcast:
			// Publicar en Redis. La distribución local ocurrirá en listenRedis
			h.publishToRedis(message)
		}
	}
}

func (h *Hub) publishToRedis(msg Mensaje) {
	if h.RedisClient == nil {
		return
	}

	// Persistir solo mensajes de chat y multimedia en MongoDB
	if msg.Tipo == "chat" || msg.Tipo == "multimedia" {
		collection := db.GetCollection("mensajes")
		_, err := collection.InsertOne(h.ctx, msg)
		if err != nil {
			log.Printf("Error persistiendo mensaje: %v", err)
		}
	}

	payload, _ := json.Marshal(msg)
	h.RedisClient.Publish(h.ctx, "chat_messages", payload)
}

func (h *Hub) cargarHistorialSala(client *Cliente) {
	collection := db.GetCollection("mensajes")

	// Obtener los últimos 50 mensajes de esta sala ordenados por tiempo
	opts := options.Find().SetLimit(50).SetSort(bson.M{"timestamp": 1})
	cursor, err := collection.Find(h.ctx, bson.M{"sala_id": client.SalaId}, opts)
	if err != nil {
		log.Printf("Error cargando historial de sala %s: %v", client.SalaId, err)
		return
	}
	defer cursor.Close(h.ctx)

	for cursor.Next(h.ctx) {
		var msg Mensaje
		if err := cursor.Decode(&msg); err == nil {
			// Enviar directamente al canal de envío del cliente
			client.Envio <- msg
		}
	}
}

func (h *Hub) listenRedis() {
	if h.RedisClient == nil {
		return
	}
	pubsub := h.RedisClient.Subscribe(h.ctx, "chat_messages")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var message Mensaje
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Error deserializando mensaje de Redis: %v", err)
			continue
		}

		h.mutex.RLock()
		if clients, ok := h.Salas[message.SalaID]; ok {
			for client := range clients {
				select {
				case client.Envio <- message:
				default:
				}
			}
		}
		h.mutex.RUnlock()
	}
}

func (h *Hub) notifyUserList(roomID string) {
	// Obtener lista global de usuarios de MongoDB
	collection := db.GetCollection("usuarios")
	cursor, err := collection.Find(h.ctx, bson.M{"sala_id": roomID, "activo": true})
	if err != nil {
		log.Printf("Error obteniendo usuarios de sala %s: %v", roomID, err)
		return
	}
	defer cursor.Close(h.ctx)

	var users []string
	for cursor.Next(h.ctx) {
		var u struct {
			Nickname string `bson:"nickname"`
		}
		if err := cursor.Decode(&u); err == nil {
			users = append(users, u.Nickname)
		}
	}

	userListJSON, _ := json.Marshal(map[string]interface{}{
		"type":  "user_list",
		"users": users,
	})

	h.publishToRedis(Mensaje{
		Tipo:   "user_list",
		Texto:  string(userListJSON),
		SalaID: roomID,
	})
}

func (h *Hub) GetRoomUsers(roomID string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0)
	if clients, ok := h.Salas[roomID]; ok {
		for client := range clients {
			users = append(users, client.Nickname)
		}
	}
	return users
}

func (h *Hub) GetRoomUserCount(roomID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.Salas[roomID]; ok {
		return len(clients)
	}
	return 0
}
