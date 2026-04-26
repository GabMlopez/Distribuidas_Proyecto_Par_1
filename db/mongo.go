package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cliente *mongo.Client
var db *mongo.Database

func ConnectDB() {
	uri := os.Getenv("MONGODB_URI")

	if uri == "" {
		log.Fatal("MONGODB_URI no está configurado en las variables de entorno")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	clienteOptions := options.Client().ApplyURI(uri)

	var err error
	cliente, err = mongo.Connect(ctx, clienteOptions)

	if err != nil {
		log.Fatal("Error durante la conexion: ", err)
	}

	err = cliente.Ping(ctx, nil)

	if err != nil {
		log.Fatal("Error al intentar hacer ping: ", err)
	}

	db = cliente.Database("chat_distribuido")
	log.Println("Conexión con Mongo Establecida")

}

func GetCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

func DisconnectDB() {
	if cliente != nil {
		cliente.Disconnect(context.Background())
	}
}
