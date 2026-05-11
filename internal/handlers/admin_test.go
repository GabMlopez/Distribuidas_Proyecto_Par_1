package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUpdateSalaHandler_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PUT("/admin/rooms/:roomId", UpdateSalaHandler)

	req, _ := http.NewRequest(http.MethodPut, "/admin/rooms/ROOM1", bytes.NewBufferString("{}"))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 when no auth claim is present, got %d", resp.Code)
	}
}

func TestDeleteSalaHandler_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/admin/rooms/:roomId", DeleteSalaHandler)

	req, _ := http.NewRequest(http.MethodDelete, "/admin/rooms/ROOM1", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 when no auth claim is present, got %d", resp.Code)
	}
}

func TestGetAllSalasAdmin_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/admin/rooms", GetAllSalasAdmin)

	req, _ := http.NewRequest(http.MethodGet, "/admin/rooms", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 when no auth claim is present, got %d", resp.Code)
	}
}

func TestGenerateRoomID(t *testing.T) {
	id := generateRoomID()
	if len(id) != 8 {
		t.Errorf("Expected room ID length 8, got %d", len(id))
	}
}
