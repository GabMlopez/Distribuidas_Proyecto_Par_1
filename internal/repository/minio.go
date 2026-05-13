package repository

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var MinioBucket string

// ConnectMinio inicializa la conexión con MinIO y asegura que el bucket exista.
func ConnectMinio() {
	// 1. Determinar el endpoint según el entorno
	endpoint := getMinIOEndpoint()
	if endpoint == "" {
		log.Println("⚠️  MINIO_ENDPOINT no configurado, MinIO no se inicializará")
		return
	}

	// 2. Credenciales
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		accessKeyID = "minioadmin"
	}

	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "minioadmin"
	}

	// 3. SSL (por defecto false en desarrollo, se puede configurar)
	useSSL := false
	if os.Getenv("MINIO_USE_SSL") == "true" {
		useSSL = true
	}

	// 4. Bucket name
	MinioBucket = os.Getenv("MINIO_BUCKET")
	if MinioBucket == "" {
		MinioBucket = "chat-uploads"
	}

	log.Printf("Conectando a MinIO en: %s (SSL: %v)", endpoint, useSSL)

	// 5. Inicializar cliente
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		// Timeout para evitar bloqueos
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
		},
	})
	if err != nil {
		log.Printf("⚠️  Error al conectar con MinIO: %v", err)
		log.Println("Continuando sin MinIO (funcionalidad de archivos limitada)")
		return
	}

	MinioClient = client

	// 6. Verificar/Crear bucket (con timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := MinioClient.BucketExists(ctx, MinioBucket)
	if err != nil {
		log.Printf("⚠️  Error verificando bucket de MinIO: %v", err)
		log.Println("Continuando sin MinIO (funcionalidad de archivos limitada)")
		MinioClient = nil
		return
	}

	if !exists {
		err = MinioClient.MakeBucket(ctx, MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("⚠️  Error creando bucket de MinIO: %v", err)
			log.Println("Continuando sin MinIO (funcionalidad de archivos limitada)")
			MinioClient = nil
			return
		}
		log.Printf("✅ Bucket '%s' creado exitosamente en MinIO", MinioBucket)

		// Configurar política pública para lectura (opcional)
		setPublicPolicy(MinioBucket)
	}

	log.Printf("✅ Conexión con MinIO Establecida en %s, bucket: %s", endpoint, MinioBucket)
}

// getMinIOEndpoint determina el endpoint correcto según el entorno
func getMinIOEndpoint() string {
	// Verificar si estamos en Render
	isRender := os.Getenv("RENDER") == "true" || os.Getenv("RENDER_SERVICE_NAME") != ""

	if isRender {
		// En Render, el servicio se comunica por nombre interno
		// Si MINIO_ENDPOINT_RENDER está configurado, usarlo
		if endpoint := os.Getenv("MINIO_ENDPOINT_RENDER"); endpoint != "" {
			log.Printf("Usando MINIO_ENDPOINT_RENDER: %s", endpoint)
			return endpoint
		}
		// Nombre por defecto en Render si no hay variable
		return "minio:9000"
	}

	// Desarrollo local: usar variable o localhost
	if endpoint := os.Getenv("MINIO_ENDPOINT_LOCAL"); endpoint != "" {
		log.Printf("Usando MINIO_ENDPOINT_LOCAL: %s", endpoint)
		return endpoint
	}

	// Variable genérica (compatibilidad)
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		log.Printf("Usando MINIO_ENDPOINT: %s", endpoint)
		return endpoint
	}

	// Fallback para desarrollo local
	log.Println("Usando localhost:9000 como endpoint por defecto")
	return "localhost:9000"
}

// setPublicPolicy configura política de acceso público al bucket
func setPublicPolicy(bucketName string) {
	if MinioClient == nil {
		return
	}

	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::` + bucketName + `/*"]
		}]
	}`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := MinioClient.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Printf("⚠️  Error configurando política pública del bucket: %v", err)
		// No fallamos por esto, es opcional
	} else {
		log.Printf("✅ Política pública configurada para bucket '%s'", bucketName)
	}
}

// UploadToMinIO sube un archivo a MinIO
func UploadToMinIO(filename string, filePath string, contentType string) (string, error) {
	if MinioClient == nil {
		return "", nil // MinIO no disponible
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Subir archivo
	objectName := filename
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := MinioClient.PutObject(ctx, MinioBucket, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	log.Printf("Archivo subido a MinIO: %s, tamaño: %d bytes", objectName, info.Size)

	// Generar URL pública
	endpoint := getMinIOEndpoint()
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	// Si estamos en Render, usar la URL pública del servicio
	if publicURL := os.Getenv("MINIO_PUBLIC_URL"); publicURL != "" {
		return publicURL + "/" + MinioBucket + "/" + objectName, nil
	}

	return scheme + "://" + endpoint + "/" + MinioBucket + "/" + objectName, nil
}
