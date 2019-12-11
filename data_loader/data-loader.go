package data_loader

import (
	log "github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var dataLogger = log.WithField("prefix", "FILE LOADER")

// DataLoader is an interface that defines how data is loded from a source into a AuthRegisterBackend interface store
type DataLoader interface {
	Init(conf interface{}) error
	LoadIntoStore(tap.AuthRegisterBackend) error
	Flush(tap.AuthRegisterBackend) error
}

