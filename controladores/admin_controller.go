package controladores

import (
	"chat_distribuido/db"
	"chat_distribuido/modelos"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateSalaRequest estructura para actualizar sala
type UpdateSalaRequest struct {
	Pin         *string `json:"pin,omitempty"`
	Tipo        *string `json:"tipo,omitempty"` // "texto" o "multimedia"
	Nombre      *string `json:"nombre,omitempty"`
	MaxFileSize *int64  `json:"max_file_size,omitempty"`
}

// CreateSalasHandler - Crear sala (requiere autenticación)
func CreateSalasHandler(c *gin.Context) {

	var req modelos.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Validar PIN
	if len(req.Pin) < 4 || len(req.Pin) > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El PIN debe tener entre 4 y 6 dígitos"})
		return
	}

	// Generar ID único para la sala
	salaID := generateRoomID()

	// Configurar tamaño máximo de archivo
	maxFileSize := req.MaxFileSize
	if maxFileSize == 0 {
		maxFileSize = 10 * 1024 * 1024 // 10MB
	}

	// Crear sala
	sala := modelos.Sala{
		SalaID:      salaID,
		Pin:         req.Pin,
		Tipo:        req.Tipo,
		MaxFileSize: maxFileSize,
		Nombre:      req.Nombre,
	}

	// Guardar en MongoDB
	collection := db.GetCollection("salas")
	_, err := collection.InsertOne(c.Request.Context(), sala)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando la sala"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sala_id":       salaID,
		"nombre":        req.Nombre,
		"pin":           req.Pin,
		"tipo":          req.Tipo,
		"max_file_size": maxFileSize,
		"message":       "Sala creada exitosamente",
	})
}

// Actualizar sala existente
func UpdateSalaHandler(c *gin.Context) {
	// Verificar autenticación
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de sala no proporcionado"})
		return
	}

	var req UpdateSalaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Buscar la sala
	collection := db.GetCollection("salas")
	var sala modelos.Sala
	err := collection.FindOne(c.Request.Context(), bson.M{"sala_id": roomID}).Decode(&sala)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando sala"})
		}
		return
	}

	// Construir objeto de actualización
	update := bson.M{}
	if req.Pin != nil {
		if len(*req.Pin) < 4 || len(*req.Pin) > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El PIN debe tener entre 4 y 6 dígitos"})
			return
		}
		update["pin"] = *req.Pin
	}
	if req.Tipo != nil {
		if *req.Tipo != "texto" && *req.Tipo != "multimedia" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo debe ser 'texto' o 'multimedia'"})
			return
		}
		update["tipo"] = *req.Tipo
	}
	if req.MaxFileSize != nil {
		if *req.MaxFileSize < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MaxFileSize no puede ser negativo"})
			return
		}
		update["max_file_size"] = *req.MaxFileSize
	}

	// Agregar metadata de actualización
	update["updated_at"] = time.Now()
	update["updated_by"] = adminID.(string)

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se proporcionaron campos para actualizar"})
		return
	}

	// Actualizar la sala
	result, err := collection.UpdateOne(
		c.Request.Context(),
		bson.M{"sala_id": roomID},
		bson.M{"$set": update},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando sala"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada o sin cambios"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sala actualizada exitosamente",
		"updated": update,
	})
}

// DeleteSalaHandler - Eliminar sala (requiere autenticación)
func DeleteSalaHandler(c *gin.Context) {
	// Verificar autenticación
	_, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de sala no proporcionado"})
		return
	}

	// Verificar si hay usuarios conectados en la sala
	userCount := hub.GetRoomUserCount(roomID)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "No se puede eliminar la sala porque tiene usuarios conectados",
			"users": userCount,
		})
		return
	}

	// Eliminar la sala
	collection := db.GetCollection("salas")
	result, err := collection.DeleteOne(c.Request.Context(), bson.M{"sala_id": roomID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando sala"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada"})
		return
	}

	// Eliminar también todos los usuarios de la sala en la BD
	usuariosCollection := db.GetCollection("usuarios")
	usuariosCollection.DeleteMany(c.Request.Context(), bson.M{"sala_id": roomID})

	c.JSON(http.StatusOK, gin.H{
		"message": "Sala eliminada exitosamente",
	})
}

// GetAllSalasAdmin - Obtener todas las salas (con PIN incluido para admin)
func GetAllSalasAdmin(c *gin.Context) {
	// Verificar autenticación
	_, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	collection := db.GetCollection("salas")
	cursor, err := collection.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo salas"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var salas []modelos.Sala
	if err = cursor.All(c.Request.Context(), &salas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decodificando salas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"salas": salas,
		"count": len(salas),
	})
}

func generateRoomID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return strings.ToUpper(hex.EncodeToString(bytes))
}
