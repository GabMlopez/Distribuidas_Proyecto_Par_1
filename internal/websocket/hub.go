package websocket

import (
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/utils"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
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
		RedisClient: repository.GetRedisClient(),
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

			// Marcar como activo en MongoDB (útil si vienen de un refresh)
			collection := repository.GetCollection("usuarios")
			if _, err := collection.UpdateOne(h.ctx,
				bson.M{"device_id": client.DeviceId, "sala_id": client.SalaId},
				bson.M{"$set": bson.M{"activo": true, "last_active": time.Now()}}); err != nil {
				log.Printf("Error updating user status in MongoDB: %v", err)
			}

			// Registrar en Redis la sesión activa del dispositivo (por DeviceID) con caducidad de seguridad (24h)
			h.RedisClient.Set(h.ctx, "device_active_session:"+client.DeviceId, client.Nickname+"|"+client.SalaId, 24*time.Hour)
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

		case client := <-h.Desregistro:
			h.mutex.Lock()
			if clients, ok := h.Salas[client.SalaId]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Envio)

					// Si la sala queda vacía, la eliminamos del mapa en memoria
					if len(clients) == 0 {
						delete(h.Salas, client.SalaId)
					}
				}
			}
			h.mutex.Unlock()

			// === GRACE PERIOD: esperar antes de marcar inactivo ===
			// Esto permite que un refresh reconecte antes de limpiar la sesión.
			go func(c *Cliente) {
				time.Sleep(5 * time.Second)

				// Verificar si el usuario se reconectó durante la espera
				h.mutex.RLock()
				reconnected := false
				if clients, ok := h.Salas[c.SalaId]; ok {
					for activeClient := range clients {
						if activeClient.DeviceId == c.DeviceId && activeClient.Nickname == c.Nickname {
							reconnected = true
							break
						}
					}
				}
				h.mutex.RUnlock()

				if reconnected {
					log.Printf("Cliente %s se reconectó a sala %s (refresh detectado, cancelando limpieza)", c.Nickname, c.SalaId)
					return
				}

				// No se reconectó → limpiar sesión de verdad
				collection := repository.GetCollection("usuarios")
				_, err := collection.UpdateOne(h.ctx,
					bson.M{"device_id": c.DeviceId, "nickname": c.Nickname},
					bson.M{
						"$set": bson.M{
							"activo":      false,
							"left_at":     time.Now(),
							"last_active": time.Now(),
						},
					})
				if err != nil {
					log.Printf("Error marcando usuario %s como inactivo: %v", c.Nickname, err)
				}

				// Verificar si el DeviceID aún tiene otra conexión activa antes de borrar Redis
				h.mutex.RLock()
				deviceStillActive := false
				for _, clients := range h.Salas {
					for activeClient := range clients {
						if activeClient.DeviceId == c.DeviceId {
							deviceStillActive = true
							break
						}
					}
					if deviceStillActive {
						break
					}
				}
				h.mutex.RUnlock()

				if !deviceStillActive {
					h.RedisClient.Del(h.ctx, "device_active_session:"+c.DeviceId)
				}

				// Notificar usuarios actualizados
				h.notifyUserList(c.SalaId)

				// Notificar que salió
				h.Broadcast <- Mensaje{
					Tipo:     "leave",
					Nickname: "Sistema",
					Texto:    c.Nickname + " ha salido de la sala",
					SalaID:   c.SalaId,
				}
				log.Printf("Cliente %s desconectado de sala %s (limpieza completa)", c.Nickname, c.SalaId)
			}(client)

		case message := <-h.Broadcast:
			// Guardar en MongoDB los mensajes que representan contenido (chat o multimedia)
			if message.Tipo == "chat" || message.Tipo == "multimedia" {
				go func(msg Mensaje) {
					collection := repository.GetCollection("mensajes")
					if msg.Timestamp == 0 {
						msg.Timestamp = time.Now().Unix()
					}
					
					// Cifrar el texto antes de guardar en la DB
					if msg.Texto != "" {
						msg.Texto = utils.EncryptMessage(msg.Texto)
					}

					_, err := collection.InsertOne(context.Background(), msg)
					if err != nil {
						log.Printf("Error persistiendo mensaje en MongoDB: %v", err)
					}
				}(message)
			}

			// Publicar en Redis. La distribución local ocurrirá en listenRedis
			h.publishToRedis(message)
		}
	}
}

func (h *Hub) publishToRedis(msg Mensaje) {
	if h.RedisClient == nil {
		return
	}
	payload, _ := json.Marshal(msg)
	h.RedisClient.Publish(h.ctx, "chat_messages", payload)
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
	collection := repository.GetCollection("usuarios")
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

// IsSessionBlocked verifica si un DeviceId, Nickname o IP ya tienen una conexión
// WebSocket activa en cualquier sala. Esto bloquea:
// - Misma pestaña/incógnito (mismo DeviceID por Canvas Fingerprint)
// - Diferente navegador en la misma máquina (misma IP local de red)
// - Mismo nickname desde cualquier lugar
// En red local, cada dispositivo físico tiene IP única (192.168.100.X),
// así que bloquear por IP NO afecta a otros dispositivos en la misma red.
func (h *Hub) IsSessionBlocked(deviceId string, nickname string, ip string) bool {
	// Excepción para pruebas de carga k6
	if len(nickname) >= 6 && nickname[:6] == "K6_VU_" {
		return false
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, clients := range h.Salas {
		for client := range clients {
			if client.DeviceId == deviceId || client.Nickname == nickname || client.Ip == ip {
				return true
			}
		}
	}
	return false
}
