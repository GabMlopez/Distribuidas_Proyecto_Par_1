package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Usuario struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UsuarioID  string             `bson:"usuario_id" json:"usuario_id"`
	Nickname   string             `bson:"nickname" json:"nickname"`
	SalaID     string             `bson:"sala_id" json:"sala_id"`
	DeviceID   string             `bson:"device_id" json:"device_id"`
	IP         string             `bson:"ip" json:"ip"`
	Activo     bool               `bson:"activo" json:"activo"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	LastActive time.Time          `bson:"last_active" json:"last_active"`
}
