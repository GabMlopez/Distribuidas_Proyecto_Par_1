package handlers

import (
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/models"
	"chat_distribuido/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type LoginRequest struct {
	Usuario     string `json:"usuario" binding:"required"`
	Contrasenia string `json:"contrasenia" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	UsuarioID string `json:"usuario_id"`
	Nickname  string `json:"nickname"`
	IsAdmin   bool   `json:"is_admin"`
}

func Login_handler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Buscar administrador en MongoDB
	var admin models.Administrador
	collection := repository.GetCollection("administradores")
	err := collection.FindOne(c.Request.Context(), bson.M{"usuario": req.Usuario}).Decode(&admin)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Verificar contraseña
	if err := utils.Check_contrasenia(req.Contrasenia, admin.Contrasenia); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Generar token para admin
	token, err := utils.GenerarTokenAdmin(admin.AdministradorID, admin.Usuario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando token"})
		return
	}

	adminUsuarioID := "admin_" + admin.AdministradorID

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		UsuarioID: adminUsuarioID,
		Nickname:  "Admin",
		IsAdmin:   true,
	})
}

// Logout maneja el cierre de sesión
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada correctamente"})
}
