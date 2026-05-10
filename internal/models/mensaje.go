package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Mensaje struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MensajeID string             `bson:"mensaje_id,omitempty" json:"mensaje_id,omitempty"`
	Nickname  string             `bson:"nickname" json:"nickname"`
	SalaID    string             `bson:"sala_id" json:"sala_id"`
	Tipo      string             `bson:"tipo" json:"tipo"`
	Timestamp int64              `bson:"timestamp" json:"timestamp"`
	FileURL   string             `bson:"file_url" json:"file_url,omitempty"`
	Texto     string                 `bson:"texto" json:"texto"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}
