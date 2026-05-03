package controladores

import (
	"chat_distribuido/controladores/sockets"
	"chat_distribuido/db"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

const (
	maxFileSize = 100 << 20 // Aumentado a 100 MB para PDFs/Documentos pesados
)

func UploadFileHandler(c *gin.Context) {
	// Obtener sala_id del contexto o query
	salaID := c.PostForm("sala_id")
	if salaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sala_id es requerido"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error al obtener el archivo"})
		return
	}
	defer file.Close()

	// Validar tamaño
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Archivo muy grande. Máximo %d MB", maxFileSize/(1024*1024))})
		return
	}

	// Validar tipo de archivo (opcional)
	allowedTypes := map[string]bool{
		// Imágenes
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".svg":  true,
		".webp": true,
		// Documentos
		".pdf":  true,
		".txt":  true,
		".docx": true,
		".doc":  true,
		".xlsx": true,
		".xls":  true,
		".pptx": true,
		".ppt":  true,
		".csv":  true,
		// Audio y Video
		".mp3":  true,
		".wav":  true,
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
		// Comprimidos
		".zip": true,
		".rar": true,
		".7z":  true,
		".tar": true,
		".gz":  true,
		// Otros datos
		".json": true,
		".xml":  true,
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	log.Printf("Recibido archivo: %s con extensión: %s", header.Filename, ext)
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de archivo no permitido"})
		return
	}

	// Generar nombre único mitigando Path Traversal
	safeHeaderName := filepath.Base(header.Filename)
	filename := fmt.Sprintf("%s_%s", salaID, safeHeaderName)

	// Obtener tamaño del archivo y tipo de contenido
	fileSize := header.Size
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Subir archivo a MinIO
	_, err = db.MinioClient.PutObject(c.Request.Context(), db.MinioBucket, filename, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("Error subiendo archivo a MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar archivo en el almacenamiento en la nube"})
		return
	}

	fileURL := fmt.Sprintf("/upload/file/%s", filename)

	// Emitir evento por WebSocket si el hub está disponible
	if hub != nil {
		hub.Broadcast <- sockets.Mensaje{
			Tipo:      "multimedia",
			Nickname:  "Sistema",
			Texto:     fmt.Sprintf("Nuevo archivo compartido: %s", header.Filename),
			SalaID:    salaID,
			FileURL:   fileURL,
			Timestamp: time.Now().Unix(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Archivo subido exitosamente",
		"filename": filename,
		"url":      fileURL,
	})
}

func GetFileHandler(c *gin.Context) {
	filename := c.Param("filename")
	
	// Mitigar Path Traversal: obtener solo el nombre base
	safeFilename := filepath.Base(filename)

	// Obtener el archivo desde MinIO
	object, err := db.MinioClient.GetObject(c.Request.Context(), db.MinioBucket, safeFilename, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error recuperando archivo"})
		return
	}
	defer object.Close()

	// Obtener información del objeto para saber su tamaño y Content-Type
	objInfo, err := object.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archivo no encontrado en MinIO"})
		return
	}

	c.DataFromReader(http.StatusOK, objInfo.Size, objInfo.ContentType, object, map[string]string{
		"Content-Disposition": fmt.Sprintf("inline; filename=\"%s\"", safeFilename),
	})
}
