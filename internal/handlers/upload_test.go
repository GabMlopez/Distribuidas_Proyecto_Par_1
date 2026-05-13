package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadFileHandler_NoSalaID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/upload/file", UploadFileHandler)

	req, _ := http.NewRequest(http.MethodPost, "/upload/file", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 when sala_id is missing, got %d", resp.Code)
	}
}

func TestGetFileHandler_NoFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/upload/file/:filename", GetFileHandler)

	// Test with no storage available (repository.MinioClient is nil by default in tests)
	req, _ := http.NewRequest(http.MethodGet, "/upload/file/test.txt", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when storage is unavailable, got %d", resp.Code)
	}
}
