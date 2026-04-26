package sockets

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	Salas map[string]map[*Cliente]bool

	Registro chan *Cliente

	Desregistro chan *Cliente

	Broadcast chan Mensaje

	mutex sync.RWMutex
}

func Nuevo_Hub() *Hub {
	return &Hub{
		Salas:       make(map[string]map[*Cliente]bool),
		Registro:    make(chan *Cliente),
		Desregistro: make(chan *Cliente),
		Broadcast:   make(chan Mensaje),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Registro:
			h.mutex.Lock()
			if _, ok := h.Salas[client.SalaId]; !ok {
				h.Salas[client.SalaId] = make(map[*Cliente]bool)
			}
			h.Salas[client.SalaId][client] = true
			h.mutex.Unlock()

			h.notifyUserList(client.SalaId)
			log.Printf("Cliente %s conectado a sala %s", client.Nickname, client.SalaId)

		case client := <-h.Desregistro:
			h.mutex.Lock()
			if clients, ok := h.Salas[client.SalaId]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Envio)

					// Si la sala queda vacía, la eliminamos
					if len(clients) == 0 {
						delete(h.Salas, client.SalaId)
					}
				}
			}
			h.mutex.Unlock()

			// Notificar usuarios actualizados
			h.notifyUserList(client.SalaId)
			log.Printf("Cliente %s desconectado de sala %s", client.Nickname, client.SalaId)

		case message := <-h.Broadcast:
			h.mutex.RLock()
			if clients, ok := h.Salas[message.SalaID]; ok {
				// Enviar mensaje a todos los clientes en la sala usando goroutines
				for client := range clients {
					select {
					case client.Envio <- message:
					default:
						close(client.Envio)
						delete(clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) notifyUserList(roomID string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.Salas[roomID]; ok {
		users := make([]string, 0, len(clients))
		for client := range clients {
			users = append(users, client.Nickname)
		}

		userListJSON, _ := json.Marshal(map[string]interface{}{
			"type":  "user_list",
			"users": users,
		})

		for client := range clients {
			select {
			case client.Envio <- Mensaje{
				Tipo:   "user_list",
				Texto:  string(userListJSON),
				SalaID: roomID,
			}:
			default:
			}
		}
	}
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
