package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadFileHandler_Validations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Falta sala_id", func(t *testing.T) {
		router := gin.Default()
		router.POST("/upload", UploadFileHandler)

		// Petición sin sala_id
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Esperaba status 400 por falta de sala_id, obtuvo %d", resp.Code)
		}
	})

	t.Run("Falta archivo", func(t *testing.T) {
		router := gin.Default()
		router.POST("/upload", UploadFileHandler)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("sala_id", "sala-123")
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Esperaba status 400 por falta de archivo, obtuvo %d", resp.Code)
		}
	})

	t.Run("Archivo no permitido", func(t *testing.T) {
		router := gin.Default()
		router.POST("/upload", UploadFileHandler)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("sala_id", "sala-123")

		// Simular archivo .exe (no permitido)
		part, _ := writer.CreateFormFile("file", "virus.exe")
		part.Write([]byte("contenido malicioso"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Esperaba status 400 por extensión no permitida, obtuvo %d", resp.Code)
		}
	})
}
