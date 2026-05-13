package main

import (
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/models"
	"chat_distribuido/internal/utils"
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("Archivo .env no encontrado")
	}

	// Conectar a MongoDB
	repository.ConnectDB()
	defer repository.DisconnectDB()

	collection := repository.GetCollection("administradores")

	// Buscar el administrador existente
	var admin models.Administrador
	err := collection.FindOne(context.Background(), bson.M{"usuario": "admin"}).Decode(&admin)

	if err == mongo.ErrNoDocuments {
		log.Println("Administrador no encontrado, creando uno nuevo...")

		// Encriptar contraseña
		hashedPassword, err := utils.Hash_contrasenia("admin1234")
		if err != nil {
			log.Fatal("Error encriptando contraseña:", err)
		}

		newAdmin := models.Administrador{
			AdministradorID: "A001",
			Usuario:         "admin",
			Contrasenia:     hashedPassword,
		}

		result, err := collection.InsertOne(context.Background(), newAdmin)
		if err != nil {
			log.Fatal("Error creando admin:", err)
		}
		log.Printf("Administrador creado con ID: %v", result.InsertedID)

	} else if err != nil {
		log.Fatal("Error buscando admin:", err)
	} else {
		// Actualizar contraseña existente
		hashedPassword, err := utils.Hash_contrasenia("admin1234")
		if err != nil {
			log.Fatal("Error encriptando contraseña:", err)
		}

		update := bson.M{
			"$set": bson.M{
				"contrasenia": hashedPassword,
			},
		}

		result, err := collection.UpdateOne(context.Background(), bson.M{"usuario": "admin"}, update)
		if err != nil {
			log.Fatal("Error actualizando contraseña:", err)
		}

		log.Printf("Contraseña actualizada para admin. Modificados: %v", result.ModifiedCount)
	}

	log.Println("¡Listo! Ahora puedes iniciar sesión con: usuario='admin', contraseña='admin1234'")
}
