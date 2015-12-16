package main

import (
	"encoding/json"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"io/ioutil"
)

type DataLoader interface {
	Init(conf interface{}) error
	LoadIntoStore(tap.AuthRegisterBackend) error
}

type FileLoaderConf struct {
	FileName string
}

type FileLoader struct {
	config FileLoaderConf
}

func (f *FileLoader) Init(conf interface{}) error {
	f.config = conf.(FileLoaderConf)
	return nil
}

func (f *FileLoader) LoadIntoStore(store tap.AuthRegisterBackend) error {
	thisSet, err := ioutil.ReadFile(f.config.FileName)
	profiles := []tap.Profile{}
	if err != nil {
		log.Error("[FILE LOADER] Load failure: ", err)
		return err
	} else {
		jsErr := json.Unmarshal(thisSet, &profiles)
		if jsErr != nil {
			log.Error("[FILE LOADER] Couldn't unmarshal profile set: ", jsErr)
			return err
		}
	}

	var loaded int
	for _, profile := range profiles {
		inputErr := AuthConfigStore.SetKey(profile.ID, profile)
		if inputErr != nil {
			log.Error("Couldn't encode configuration: ", inputErr)
		} else {
			loaded += 1
		}
	}

	log.Info("[FILE LOADER] Loaded: ", loaded, " profiles...")
	return nil
}
