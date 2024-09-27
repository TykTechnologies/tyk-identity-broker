package model

type DBObject interface {
	GetObjectID() ObjectID
	SetObjectID(id ObjectID)
	TableName() string
}
