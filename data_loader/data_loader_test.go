package data_loader

import (
	"github.com/TykTechnologies/storage/persistent"
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk-identity-broker/configuration"
)

func TestCreateDataMongoLoader(t *testing.T) {
	isMongo, url, driver := isMongoEnv()
	t.Skip(!isMongo)
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

	isMongo, url, driver := isMongoEnv()
	t.Skip(!isMongo)
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

func isMongoEnv() (bool, string, string) {
	storageType := os.Getenv("TYK_IB_STORAGE_STORAGETYPE")
	mongoURL := os.Getenv("TYK_IB_STORAGE_MONGOCONF_MONGOURL")
	mongoDriver := os.Getenv("TYK_IB_STORAGE_MONGOCONF_DRIVER")

	// Check if MongoDB is the storage type and both URL and driver are set
	if storageType == "mongo" && mongoURL != "" && mongoDriver != "" {
		return true, mongoURL, mongoDriver
	}

	return false, "", ""
}
