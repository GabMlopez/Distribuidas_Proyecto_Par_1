package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type Cliente struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Envio    chan Mensaje
	Nickname string
	SalaId   string
	DeviceId string
	Ip       string
}

func (c *Cliente) ReadPump() {
	defer func() {
		c.Hub.Desregistro <- c
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing websocket connection: %v", err)
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("Error setting read deadline in pong handler: %v", err)
		}
		return nil
	})

	for {
		var message Mensaje
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if message.Tipo == "" {
			message.Tipo = "chat"
		}
		message.Nickname = c.Nickname
		message.SalaID = c.SalaId
		message.Timestamp = time.Now().Unix()

		c.Hub.Broadcast <- message
	}

}

func (c *Cliente) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing websocket connection in write pump: %v", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.Envio:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("Error setting write deadline: %v", err)
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("Error sending close message: %v", err)
				}
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("Error setting write deadline for ping: %v", err)
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
