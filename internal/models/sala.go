package modelos

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sala struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SalaID      string             `bson:"sala_id" json:"sala_id"`
	Pin         string             `bson:"pin" json:"pin"`
	Tipo        string             `bson:"tipo" json:"tipo"`
	MaxFileSize int64              `bson:"max_file_size" json:"max_file_size"`
	Nombre      string             `bson:"nombre" json:"nombre"`
}

type CreateRoomRequest struct {
	Pin         string `json:"pin" binding:"required,min=4,max=6"`
	Tipo        string `json:"tipo" binding:"required,oneof=texto multimedia"`
	MaxFileSize int64  `json:"max_file_size"`
	Nombre      string `json:"nombre"`
}
