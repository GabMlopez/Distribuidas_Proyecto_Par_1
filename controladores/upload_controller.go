package controladores

import (
	"chat_distribuido/controladores/sockets"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxFileSize = 10 << 20 // 10 MB
	uploadDir   = "./uploads"
)

func init() {
	// Crear directorio de uploads si no existe
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
}

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
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".txt":  true,
		".mp3":  true,
		".mp4":  true,
		".docx": true,
		".doc":  true,
		".xlsx": true,
		".xls":  true,
		".pptx": true,
		".ppt":  true,
		".zip":  true,
		".csv":  true,
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	log.Printf("Recibido archivo: %s con extensión: %s", header.Filename, ext)
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de archivo no permitido"})
		return
	}

	// Generar nombre único
	filename := fmt.Sprintf("%s_%s", salaID, header.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Guardar archivo
	out, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar archivo"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al copiar archivo"})
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
	filepath := filepath.Join(uploadDir, filename)

	// Verificar si el archivo existe
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archivo no encontrado"})
		return
	}

	c.File(filepath)
}
