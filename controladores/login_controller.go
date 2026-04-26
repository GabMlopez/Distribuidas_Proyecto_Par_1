package controladores

import (
	"chat_distribuido/db"
	"chat_distribuido/modelos"
	"chat_distribuido/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var jwtSecret []byte

type Login_req struct {
	Usuario     string `json:"usuario" binding:"required"`
	Contrasenia string `json:"contrasenia" binding:"required"`
}

type Login_res struct {
	Token string `json:"token"`
}

func Login_handler(c *gin.Context) {
	var req Login_req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Datos inválidos: " + err.Error(),
		})
		return
	}

	// Buscar administrador en MongoDB
	var admin modelos.Administrador
	collection := db.GetCollection("administradores")
	err := collection.FindOne(c.Request.Context(), bson.M{"usuario": req.Usuario}).Decode(&admin)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciales inválidas",
		})
		return
	}

	// Verificar contraseña usando la función del paquete utils
	if err := utils.Check_contrasenia(req.Contrasenia, admin.Contrasenia); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciales inválidas",
		})
		return
	}

	// Generar token usando la función unificada
	token, err := utils.GenerarTokenAdmin(admin.AdministradorID, admin.Usuario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando token",
		})
		return
	}

	c.JSON(http.StatusOK, Login_res{
		Token: token,
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada exitosamente"})
}
