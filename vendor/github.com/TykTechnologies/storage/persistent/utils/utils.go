package utils

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2"
)

type Info struct {
	Type    DBType
	Version string
}

type DBType string

const (
	StandardMongo DBType = "mongo"
	AWSDocumentDB DBType = "docdb"
	CosmosDB      DBType = "cosmosdb"
)

func IsErrNoRows(err error) bool {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return true
	}

	if errors.Is(err, mgo.ErrNotFound) {
		return true
	}

	return false
}
