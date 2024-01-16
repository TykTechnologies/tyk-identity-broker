package data_loader

import (
	"os"
	"testing"

	"github.com/TykTechnologies/storage/persistent"
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/tyk-identity-broker/configuration"
)

func TestCreateDataMongoLoader(t *testing.T) {
	isMongo := isMongoEnv()
	t.Skip(!isMongo)

	url, driver := MongoEnvConf()

	conf := configuration.Configuration{
		Storage: &configuration.Storage{
			StorageType: driver,
			MongoConf: &configuration.MongoConf{
				MongoURL: url,
			},
		},
	}

	dataLoader, err := CreateDataLoader(conf, "")
	if err != nil {
		t.Fatalf("creating Mongo data loader: %v", err)
	}

	if _, ok := dataLoader.(*MongoLoader); !ok {
		t.Fatalf("type of data loader is not correct; expected '*data_loader.MongoLoader' but got '%T'", dataLoader)
	}
}

func TestFlush(t *testing.T) {

	isMongo := isMongoEnv()
	t.Skip(!isMongo)

	url, driver := MongoEnvConf()
	loader := MongoLoader{}

	var err error
	loader.store, err = persistent.NewPersistentStorage(&persistent.ClientOpts{
		ConnectionString: url,
		UseSSL:           false,
		Type:             driver,
	})

	assert.Nil(t, err)

	authStore := &backends.InMemoryBackend{}
	err = loader.Flush(authStore)
	assert.Nil(t, err)
}

func isMongoEnv() bool {
	storageType := os.Getenv("TYK_IB_STORAGE_STORAGETYPE")

	// Check if MongoDB is the storage type and both URL and driver are set
	return storageType == "mongo"
}

func MongoEnvConf() (string, string) {
	mongoURL := os.Getenv("TYK_IB_STORAGE_MONGOCONF_MONGOURL")
	mongoDriver := os.Getenv("TYK_IB_STORAGE_MONGOCONF_DRIVER")

	return mongoURL, mongoDriver

}
