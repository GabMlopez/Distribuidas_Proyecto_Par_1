package db

import (
	"chat_distribuido/modelos"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GuardarMensaje guarda un mensaje en MongoDB
func GuardarMensaje(mensaje modelos.Mensaje) error {
	collection := GetCollection("mensajes")

	// Establecer ID si no tiene
	if mensaje.Id.IsZero() {
		mensaje.Id = primitive.NewObjectID()
	}

	// Establecer timestamp si no tiene
	if mensaje.Timestamp == 0 {
		mensaje.Timestamp = time.Now().Unix()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, mensaje)
	if err != nil {
		log.Printf("Error guardando mensaje en DB: %v", err)
		return err
	}

	return nil
}

// ObtenerHistorialMensajes obtiene los últimos mensajes de una sala
func ObtenerHistorialMensajes(salaID string, limit int64, offset int64) ([]modelos.Mensaje, error) {
	collection := GetCollection("mensajes")

	// Opciones de búsqueda
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"timestamp": -1}) // Más recientes primero
	findOptions.SetLimit(limit)
	findOptions.SetSkip(offset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"sala_id": salaID}, findOptions)
	if err != nil {
		log.Printf("Error obteniendo mensajes: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []modelos.Mensaje
	if err = cursor.All(ctx, &messages); err != nil {
		log.Printf("Error decodificando mensajes: %v", err)
		return nil, err
	}

	// Invertir para orden cronológico
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CrearIndicesMensajes crea índices para la colección de mensajes
func CrearIndicesMensajes() {
	collection := GetCollection("mensajes")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice para búsqueda por sala_id y timestamp
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sala_id", Value: 1},
			{Key: "timestamp", Value: -1},
		},
		Options: options.Index().SetName("idx_sala_timestamp"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Error creando índice para mensajes: %v", err)
	} else {
		log.Println("Índice para mensajes creado exitosamente")
	}
}
