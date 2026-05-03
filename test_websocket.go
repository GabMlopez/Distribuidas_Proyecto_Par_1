package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"chat_distribuido/db"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func TestWS() {
	_ = godotenv.Load()
	db.ConnectRedis()
	db.ConnectDB()

	// Esperar que main.go arranque y limpie
	time.Sleep(3 * time.Second)

	// Inyectar usuario en MongoDB para que pase la validación
	col := db.GetCollection("usuarios")
	col.InsertOne(context.Background(), bson.M{"nickname": "PruebaUser", "sala_id": "sala1", "activo": true})
	col.InsertOne(context.Background(), bson.M{"nickname": "PruebaUser", "sala_id": "sala2", "activo": true})

	// Limpieza al final
	defer func() {
		col.DeleteMany(context.Background(), bson.M{"nickname": "PruebaUser"})
		db.DisconnectDB()
	}()

	time.Sleep(1 * time.Second)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws/sala1", RawQuery: "nickname=PruebaUser&sala_id=sala1"}
	log.Printf("Conectando a %s", u.String())

	// Conexión 1
	c1, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Error en conexión 1: %v\n", err)
	}
	defer c1.Close()
	log.Println("Conexión 1: ÉXITO (Usuario en sala1)")

	// Intentar conectar el mismo user a OTRA sala (sala2)
	time.Sleep(1 * time.Second) // Dar un segundo para que el backend mande log de que entró

	u2 := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws/sala2", RawQuery: "nickname=PruebaUser&sala_id=sala2"}
	log.Printf("Conectando a %s", u2.String())
	c2, resp2, err2 := websocket.DefaultDialer.Dial(u2.String(), nil)

	if err2 != nil {
		log.Printf("Conexión 2 falló a nivel HTTP (Status: %v). Error: %v\n", resp2.StatusCode, err2)
	} else {
		// Leer mensaje de error que devuelve el socket
		var msg map[string]interface{}
		c2.ReadJSON(&msg)
		log.Printf("=> RESULTADO DE CONEXIÓN 2: %v", msg)
		c2.Close()
	}
}
