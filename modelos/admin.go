package modelos

import "go.mongodb.org/mongo-driver/bson/primitive"

type Administrador struct {
	Id              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdministradorID string             `bson:"administrador_id" json:"administrador_id"`
	Usuario         string             `bson:"usuario" json:"usuario"`
	Contrasenia     string             `bson:"contrasenia" json:"contrasenia"`
}
