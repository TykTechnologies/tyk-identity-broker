//go:build test_mongo
// +build test_mongo

package data_loader_test

import (
	"testing"

	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	"github.com/TykTechnologies/tyk-identity-broker/data_loader"
)

func TestCreateDataMongoLoader(t *testing.T) {
	conf := configuration.Configuration{
		Storage: &configuration.Storage{
			StorageType: configuration.MONGO,
			MongoConf: &configuration.MongoConf{
				MongoURL: "mongodb://tyk-mongo:27017/tyk_tib",
			},
		},
	}

	dataLoader, err := data_loader.CreateDataLoader(conf, nil)
	if err != nil {
		t.Fatalf("creating Mongo data loader: %v", err)
	}

	if _, ok := dataLoader.(*data_loader.MongoLoader); !ok {
		t.Fatalf("type of data loader is not correct; expected '*data_loader.MongoLoader' but got '%T'", dataLoader)
	}
}
