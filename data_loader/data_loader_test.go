package data_loader

import (
	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	"reflect"
	"testing"
)

func TestCreateDataMongoLoader(t *testing.T){

	conf := configuration.Configuration{
		Storage:               &configuration.Storage{
			StorageType:       configuration.MONGO,
			MongoConf:         &configuration.MongoConf{
				MongoURL:                   "mongodb://tyk-mongo:27017/tyk_tib",
			},
		},
	}
	dataLoader, err := CreateDataLoader(conf, nil)

	if err != nil {
		t.Error("creating mongo data loader: "+err.Error())
	}

	loaderType := reflect.TypeOf(dataLoader)
	if loaderType.String() != "*data_loader.MongoLoader"{
		t.Error("type of data loader is not correct. Expected *data_loader.MongoLoader but get:"+loaderType.String())
	}
}
