package db

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var MinioBucket = "chat-uploads"

// ConnectMinio inicializa la conexión con MinIO y asegura que el bucket exista.
func ConnectMinio() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9191" // Valor por defecto
	}
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		accessKeyID = "admin"
	}
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "password123"
	}

	useSSL := false // Usar false en desarrollo local sin HTTPS

	// Inicializar el objeto cliente de MinIO
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Error al conectar con MinIO: %v", err)
	}

	MinioClient = client

	// Comprobar si el bucket especificado ya existe
	ctx := context.Background()
	exists, errBucketExists := MinioClient.BucketExists(ctx, MinioBucket)
	if errBucketExists != nil {
		log.Fatalf("Error comprobando bucket de MinIO: %v", errBucketExists)
	}
	if !exists {
		// Crear el bucket
		err = MinioClient.MakeBucket(ctx, MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Error creando bucket de MinIO: %v", err)
		}
		log.Printf("Bucket %s creado exitosamente en MinIO\n", MinioBucket)
	}

	log.Println("Conexión con MinIO Establecida")
}
