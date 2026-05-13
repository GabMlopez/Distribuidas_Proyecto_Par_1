package handlers

import (
	"chat_distribuido/internal/websocket"
	"chat_distribuido/internal/repository"
	"chat_distribuido/internal/models"
	"chat_distribuido/internal/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var hub *websocket.Hub

func SetHub(h *websocket.Hub) {
	hub = h
}

type JoinRoomRequest struct {
	SalaID   string `json:"sala_id" binding:"required"`
	Pin      string `json:"pin" binding:"required"`
	Nickname string `json:"nickname"`
	DeviceID string `json:"device_id" binding:"required"`
}

type JoinRoomResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	UsuarioID    string `json:"usuario_id,omitempty"`
	Tipo         string `json:"tipo,omitempty"`
	Token        string `json:"token,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	IsNewUser    bool   `json:"is_new_user"`
	PreviousRoom string `json:"previous_room,omitempty"`
}

type RoomResponse struct {
	SalaID   string `json:"sala_id"`
	Tipo     string `json:"tipo"`
	Usuarios int    `json:"usuarios"`
	Nombre   string `json:"nombre,omitempty"`
}

func validarSalaYPin(ctx *gin.Context, salaID, pin string) (*models.Sala, error) {
	var sala models.Sala
	collection := repository.GetCollection("salas")
	err := collection.FindOne(ctx.Request.Context(), bson.M{"sala_id": salaID}).Decode(&sala)
	if err != nil {
		return nil, err
	}

	if sala.Pin != pin {
		return nil, mongo.ErrNoDocuments
	}

	return &sala, nil
}

func verificarCapacidadSala(salaID string) error {
	if hub == nil {
		return nil // No podemos verificar sin hub
	}

	userCount := hub.GetRoomUserCount(salaID)
	if userCount >= 50 {
		return &CapacityError{Message: "Sala llena (máximo 50 usuarios)"}
	}
	return nil
}

type CapacityError struct {
	Message string
}

func (e *CapacityError) Error() string {
	return e.Message
}

// getRealIP obtiene la IP real del cliente considerando proxies
func getRealIP(c *gin.Context) string {
	// Verificar X-Forwarded-For (para proxies/load balancers)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// Tomar la primera IP de la lista
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	ip := c.ClientIP()

	if ip == "::1" {
		ip = "127.0.0.1"
	}

	return ip
}

// buscarDispositivoExistente busca un dispositivo activo por su IP
func buscarDispositivoExistente(ctx *gin.Context, ip string) (*models.Usuario, error) {
	var usuario models.Usuario
	collection := repository.GetCollection("usuarios")
	err := collection.FindOne(ctx.Request.Context(), bson.M{"ip": ip, "activo": true}).Decode(&usuario)
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

// actualizarUsuarioSala actualiza un usuario existente a una nueva sala
func actualizarUsuarioSala(ctx *gin.Context, deviceID, salaID, nickname, ip string) error {
	collection := repository.GetCollection("usuarios")
	update := bson.M{
		"$set": bson.M{
			"sala_id":     salaID,
			"nickname":    nickname,
			"activo":      true,
			"ip":          ip,
			"last_active": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx.Request.Context(),
		bson.M{"device_id": deviceID},
		update)
	return err
}

func verificarNicknameDisponible(ctx *gin.Context, salaID, nickname string) bool {
	collection := repository.GetCollection("usuarios")
	var usuario models.Usuario
	err := collection.FindOne(ctx.Request.Context(), bson.M{
		"sala_id":  salaID,
		"nickname": nickname,
		"activo":   true,
	}).Decode(&usuario)
	return err != nil
}

func generarNicknameAutomatico(ctx *gin.Context, salaID string) string {
	collection := repository.GetCollection("usuarios")

	// Contar el total
	totalUsersInSystem, _ := collection.CountDocuments(ctx.Request.Context(), bson.M{})
	baseNumber := int(totalUsersInSystem) + 1

	nickname := "user" + strconv.Itoa(baseNumber)

	// Verificar si ya existe en la sala actual
	for !verificarNicknameDisponible(ctx, salaID, nickname) {
		baseNumber++
		nickname = "user" + strconv.Itoa(baseNumber)
	}

	return nickname
}

func procesarNickname(ctx *gin.Context, salaID string, nicknameSolicitado string, usuarioExistente *models.Usuario) (string, error) {
	// Caso: Usuario proporciona nickname
	if nicknameSolicitado != "" {
		if !verificarNicknameDisponible(ctx, salaID, nicknameSolicitado) {
			return "", &NicknameError{Message: "El nickname '" + nicknameSolicitado + "' ya está en uso en esta sala"}
		}
		return nicknameSolicitado, nil
	}

	// Caso: Usuario existente sin nickname proporcionado
	if usuarioExistente != nil && usuarioExistente.Nickname != "" {
		if verificarNicknameDisponible(ctx, salaID, usuarioExistente.Nickname) {
			return usuarioExistente.Nickname, nil
		}
		return generarNicknameAutomatico(ctx, salaID), nil
	}

	return generarNicknameAutomatico(ctx, salaID), nil
}

type NicknameError struct {
	Message string
}

func (e *NicknameError) Error() string {
	return e.Message
}

func crearNuevoUsuario(ctx *gin.Context, salaID, deviceID, nickname, ip string) (*models.Usuario, error) {
	collection := repository.GetCollection("usuarios")

	usuarioID := generateUserID()

	usuario := models.Usuario{
		UsuarioID: usuarioID,
		Nickname:  nickname,
		SalaID:    salaID,
		DeviceID:  deviceID,
		Activo:    true,
		IP:        ip,
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx.Request.Context(), usuario)
	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

func UnirseSalaHandler(c *gin.Context) {
	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	if hub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno: hub no inicializado"})
		return
	}

	sala, err := validarSalaYPin(c, req.SalaID, req.Pin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sala no encontrada o PIN incorrecto"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando sala"})
		}
		return
	}

	if err := verificarCapacidadSala(req.SalaID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	clientIP := getRealIP(c)

	// === CAPA 0: Triple validación en Hub (DeviceID + Nickname + IP) ===
	// En red local cada dispositivo tiene IP única → bloquea incógnito y otros navegadores
	if hub.IsSessionBlocked(req.DeviceID, req.Nickname, clientIP) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "⚠️ Bloqueo de Seguridad: Ya existe una sesión activa desde este dispositivo (IP: " + clientIP + "). No se permiten múltiples navegadores, pestañas ni modo incógnito simultáneamente.",
		})
		return
	}

	// === CAPA 1: Verificar en Redis si este DeviceID ya tiene sesión activa ===
	activeSession, errRedis := repository.RedisClient.Get(c.Request.Context(), "device_active_session:"+req.DeviceID).Result()
	if errRedis == nil && activeSession != "" {
		expectedSession := req.Nickname + "|" + req.SalaID
		if activeSession != expectedSession {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Este dispositivo ya tiene una sesión activa en otra sala o con otro usuario.",
			})
			return
		}
	}

	var usuarioActivoPorDevice models.Usuario
	errDevice := repository.GetCollection("usuarios").FindOne(c.Request.Context(),
		bson.M{"device_id": req.DeviceID, "activo": true}).Decode(&usuarioActivoPorDevice)
	if errDevice == nil {
		if usuarioActivoPorDevice.Nickname != req.Nickname || usuarioActivoPorDevice.SalaID != req.SalaID {
			c.JSON(http.StatusConflict, gin.H{
				"error": "⚠️ Sesión Duplicada Detectada: El sistema ha detectado que tu navegador (DeviceID) ya está conectado en otra sala. Solo puedes estar activo en una sala a la vez.",
			})
			return
		}
	}

	var usuarioExistente models.Usuario
	err = repository.GetCollection("usuarios").FindOne(c.Request.Context(),
		bson.M{"device_id": req.DeviceID}).Decode(&usuarioExistente)
	fmt.Printf("=== DEPURACIÓN UNIRSE SALA ===\nDeviceID: %s\nIP: %s\nError búsqueda usuario: %v\n", req.DeviceID, clientIP, err)
	var usuarioID string
	var nickname string
	var isNewUser bool
	var previousRoom string

	if err == nil {
		// Dispositivo existe
		isNewUser = false
		usuarioID = usuarioExistente.UsuarioID
		previousRoom = usuarioExistente.SalaID

		// Verificar si ya está activo en OTRA sala
		if usuarioExistente.Activo && usuarioExistente.SalaID != "" && usuarioExistente.SalaID != req.SalaID {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Ya tienes una sesión activa en la sala: " + usuarioExistente.SalaID + ". Cierra esa sesión primero.",
			})
			return
		}

		// Procesar nickname
		nickname, err = procesarNickname(c, req.SalaID, req.Nickname, &usuarioExistente)
		if err != nil {
			if nicknameErr, ok := err.(*NicknameError); ok {
				c.JSON(http.StatusConflict, gin.H{"error": nicknameErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando nickname"})
			}
			return
		}

		// Actualizar usuario
		update := bson.M{
			"$set": bson.M{
				"sala_id":     req.SalaID,
				"nickname":    nickname,
				"activo":      true,
				"ip":          getRealIP(c),
				"last_active": time.Now(),
			},
		}
		_, err = repository.GetCollection("usuarios").UpdateOne(c.Request.Context(),
			bson.M{"device_id": req.DeviceID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario"})
			return
		}

	} else if err == mongo.ErrNoDocuments {
		// Nuevo dispositivo
		isNewUser = true

		// Procesar nickname
		nickname, err = procesarNickname(c, req.SalaID, req.Nickname, nil)
		if err != nil {
			if nicknameErr, ok := err.(*NicknameError); ok {
				c.JSON(http.StatusConflict, gin.H{"error": nicknameErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando nickname"})
			}
			return
		}

		usuarioID = generateUserID()
		usuario := models.Usuario{
			UsuarioID:  usuarioID,
			Nickname:   nickname,
			SalaID:     req.SalaID,
			DeviceID:   req.DeviceID,
			Activo:     true,
			IP:         getRealIP(c),
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
		}

		_, err = repository.GetCollection("usuarios").InsertOne(c.Request.Context(), usuario)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error registrando usuario"})
			return
		}

	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando usuario"})
		return
	}

	// Generar token
	token, err := utils.GenerarTokenUser(usuarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando token"})
		return
	}

	response := JoinRoomResponse{
		Success:   true,
		UsuarioID: usuarioID,
		Tipo:      sala.Tipo,
		Token:     token,
		Nickname:  nickname,
		IsNewUser: isNewUser,
	}

	if previousRoom != "" && previousRoom != req.SalaID {
		response.PreviousRoom = previousRoom
		response.Message = "Has cambiado de sala. Anterior sala: " + previousRoom
	} else {
		response.Message = "Te has unido a la sala exitosamente"
	}

	c.JSON(http.StatusOK, response)
}
func getActiveUsersInRoom(ctx *gin.Context, salaID string) []map[string]string {
	collection := repository.GetCollection("usuarios")

	cursor, err := collection.Find(ctx.Request.Context(), bson.M{
		"sala_id": salaID,
		"activo":  true,
	})

	if err != nil {
		return []map[string]string{}
	}
	defer cursor.Close(ctx.Request.Context())

	var usuarios []models.Usuario
	if err = cursor.All(ctx.Request.Context(), &usuarios); err != nil {
		return []map[string]string{}
	}

	var usersList []map[string]string
	for _, u := range usuarios {
		usersList = append(usersList, map[string]string{
			"usuario_id": u.UsuarioID,
			"nickname":   u.Nickname,
		})
	}

	return usersList
}

func DejarSalaHandler(c *gin.Context) {
	var req struct {
		UsuarioID string `json:"usuario_id" binding:"required"`
		SalaID    string `json:"sala_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	println("=== DEPURACIÓN DEJAR SALA ===")
	println("UsuarioID recibido:", req.UsuarioID)
	println("SalaID recibida:", req.SalaID)
	println("Longitud del ID:", len(req.UsuarioID))

	collection := repository.GetCollection("usuarios")
	update := bson.M{
		"$set": bson.M{
			"activo":      false,
			"sala_id":     "",
			"left_at":     time.Now(),
			"last_active": time.Now(),
		},
	}

	result, err := collection.UpdateOne(c.Request.Context(),
		bson.M{"usuario_id": req.UsuarioID},
		update)

	// Limpiar sesión en Redis si existe
	// Requiere que busquemos primero el dispositivo del usuario para saber su DeviceID,
	// pero en 'DejarSalaHandler' solo tenemos usuario_id y sala_id.
	// Vamos a recuperar el DeviceID del usuario para limpiar correctamente Redis.
	var usuario models.Usuario
	errBuscar := repository.GetCollection("usuarios").FindOne(c.Request.Context(), bson.M{"usuario_id": req.UsuarioID}).Decode(&usuario)
	if errBuscar == nil && usuario.DeviceID != "" {
		repository.RedisClient.Del(c.Request.Context(), "device_active_session:"+usuario.DeviceID)
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "El usuario no se encuentra en esa sala"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al salir de la sala"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Has salido de la sala",
	})
}

func ActualizarNicknameHandler(c *gin.Context) {
	var req struct {
		UsuarioID     string `json:"usuario_id" binding:"required"`
		NicknameNuevo string `json:"nickname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	collection := repository.GetCollection("usuarios")
	update := bson.M{
		"$set": bson.M{
			"nickname":    req.NicknameNuevo,
			"last_active": time.Now(),
		},
	}
	result, err := collection.UpdateOne(c.Request.Context(),
		bson.M{"usuario_id": req.UsuarioID},
		update)
	if err != nil || result.ModifiedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando nickname"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Nickname actualizado exitosamente a: " + req.NicknameNuevo,
	})

}

func ObtenerUsuariosSala(c *gin.Context) {
	salaID := c.Param("roomId")

	usersList := getActiveUsersInRoom(c, salaID)

	c.JSON(http.StatusOK, gin.H{
		"sala_id": salaID,
		"count":   len(usersList),
		"users":   usersList,
	})
}

func ListaSalas(c *gin.Context) {
	if hub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}

	collection := repository.GetCollection("salas")
	opts := options.Find().SetProjection(bson.M{"pin": 0})

	cursor, err := collection.Find(c.Request.Context(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo salas"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var salas []models.Sala
	if err = cursor.All(c.Request.Context(), &salas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decodificando salas"})
		return
	}

	response := make([]RoomResponse, 0)
	for _, sala := range salas {
		response = append(response, RoomResponse{
			SalaID:   sala.SalaID,
			Tipo:     sala.Tipo,
			Usuarios: hub.GetRoomUserCount(sala.SalaID),
			Nombre:   sala.Nombre,
		})
	}

	c.JSON(http.StatusOK, response)
}

func generateUserID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GetMessagesHandler(c *gin.Context) {
	roomId := c.Param("roomId")
	if roomId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId es requerido"})
		return
	}

	collection := repository.GetCollection("mensajes")
	// Obtener los últimos 100 mensajes, ordenados por timestamp ascendente
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	findOptions.SetLimit(100)

	cursor, err := collection.Find(c.Request.Context(), bson.M{"sala_id": roomId}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo mensajes"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var mensajes []models.Mensaje
	if err = cursor.All(c.Request.Context(), &mensajes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decodificando mensajes"})
		return
	}

	// Como los obtuvimos ordenados descendentemente (para tener los más recientes),
	// los invertimos para devolverlos en orden cronológico ascendente y desciframos.
	for i, j := 0, len(mensajes)-1; i < j; i, j = i+1, j-1 {
		// Descifrar si tiene texto
		if mensajes[i].Texto != "" {
			mensajes[i].Texto = utils.DecryptMessage(mensajes[i].Texto)
		}
		if mensajes[j].Texto != "" {
			mensajes[j].Texto = utils.DecryptMessage(mensajes[j].Texto)
		}
		mensajes[i], mensajes[j] = mensajes[j], mensajes[i]
	}

	// Si hay número impar de mensajes, el del medio no se descifró en el bucle anterior
	if len(mensajes)%2 != 0 {
		mid := len(mensajes) / 2
		if mensajes[mid].Texto != "" {
			mensajes[mid].Texto = utils.DecryptMessage(mensajes[mid].Texto)
		}
	}

	if mensajes == nil {
		mensajes = make([]models.Mensaje, 0)
	}

	c.JSON(http.StatusOK, mensajes)
}
