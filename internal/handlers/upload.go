package handlers

import (
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/websocket"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"chat_distribuido/internal/models"
	"chat_distribuido/internal/utils"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/bson"
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
	if repository.MinioClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "El servicio de almacenamiento no está disponible actualmente. El archivo no se pudo subir."})
		return
	}

	_, err = repository.MinioClient.PutObject(c.Request.Context(), repository.MinioBucket, filename, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("Error subiendo archivo a MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar archivo en el almacenamiento en la nube"})
		return
	}

	fileURL := fmt.Sprintf("/upload/file/%s", filename)

	// Obtener el nickname del usuario
	usuarioID, exists := c.Get("usuario")
	nickname := "Sistema"
	if exists {
		var usuario models.Usuario
		err := repository.GetCollection("usuarios").FindOne(c.Request.Context(), bson.M{"usuario_id": usuarioID.(string)}).Decode(&usuario)
		if err == nil && usuario.Nickname != "" {
			nickname = usuario.Nickname
		}
	}

	metadata := map[string]interface{}{
		"filename":     header.Filename,
		"size":         fileSize,
		"content_type": contentType,
		"extension":    ext,
	}

	// Emitir evento por WebSocket si el hub está disponible
	if hub != nil {
		hub.Broadcast <- websocket.Mensaje{
			Tipo:      "multimedia",
			Nickname:  nickname,
			Texto:     fmt.Sprintf("Nuevo archivo compartido: %s", header.Filename),
			SalaID:    salaID,
			FileURL:   fileURL,
			Timestamp: time.Now().Unix(),
			Metadata:  metadata,
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
	if decoded, err := url.PathUnescape(filename); err == nil {
		filename = decoded
	}

	// Mitigar Path Traversal: obtener solo el nombre base
	safeFilename := filepath.Base(filename)

	if repository.MinioClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Servicio de almacenamiento no disponible"})
		return
	}

	// Obtener el archivo desde MinIO
	object, err := repository.MinioClient.GetObject(c.Request.Context(), repository.MinioBucket, safeFilename, minio.GetObjectOptions{})
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

func DeleteFileHandler(c *gin.Context) {
	filename := c.Param("filename")
	if decoded, err := url.PathUnescape(filename); err == nil {
		filename = decoded
	}
	safeFilename := filepath.Base(filename)

	usuarioID, exists := c.Get("usuario")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	var usuario models.Usuario
	err := repository.GetCollection("usuarios").FindOne(c.Request.Context(), bson.M{"usuario_id": usuarioID.(string)}).Decode(&usuario)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	fileURL := fmt.Sprintf("/upload/file/%s", safeFilename)

	// Buscar el mensaje asociado
	var mensaje models.Mensaje
	err = repository.GetCollection("mensajes").FindOne(c.Request.Context(), bson.M{"file_url": fileURL}).Decode(&mensaje)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mensaje o archivo no encontrado"})
		return
	}

	// Validar que el usuario sea el dueño
	if mensaje.Nickname != usuario.Nickname {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para eliminar esta imagen"})
		return
	}

	// Eliminar de MinIO
	if repository.MinioClient != nil {
		err = repository.MinioClient.RemoveObject(c.Request.Context(), repository.MinioBucket, safeFilename, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("Error eliminando archivo de MinIO: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el archivo físico"})
			return
		}
	}

	// Actualizar en MongoDB en lugar de borrar
	textoCifrado := utils.EncryptMessage("🚫 Este archivo multimedia fue eliminado")
	update := bson.M{
		"$set": bson.M{
			"file_url": "",
			"texto":    textoCifrado,
		},
	}
	_, err = repository.GetCollection("mensajes").UpdateOne(c.Request.Context(), bson.M{"_id": mensaje.Id}, update)
	if err != nil {
		log.Printf("Error actualizando mensaje en MongoDB: %v", err)
	}

	// Emitir evento por WebSocket para eliminar en UI
	if hub != nil {
		hub.Broadcast <- websocket.Mensaje{
			Tipo:    "delete_message",
			SalaID:  mensaje.SalaID,
			FileURL: fileURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Archivo eliminado correctamente"})
}
