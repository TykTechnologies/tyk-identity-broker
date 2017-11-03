package main

import (
	"encoding/json"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"io/ioutil"
	"path"
	"strconv"
	"time"
)

// DataLoader is an interface that defines how data is loded from a source into a AuthRegisterBackend interface store
type DataLoader interface {
	Init(conf interface{}) error
	LoadIntoStore(tap.AuthRegisterBackend) error
	Flush(tap.AuthRegisterBackend) error
}

// FileLoaderConf is the configuration struct for a FileLoader, takes a filename as main init
type FileLoaderConf struct {
	FileName string
}

// FileLoader implements DataLoader and will load TAP Profiles from a file
type FileLoader struct {
	config FileLoaderConf
}

// Init initialises the file loader
func (f *FileLoader) Init(conf interface{}) error {
	f.config = conf.(FileLoaderConf)
	return nil
}

// LoadIntoStore will load, unmarshal and copy profiles into a an AuthRegisterBackend
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

	log.Info("[FILE LOADER] Loaded: ", loaded, " profiles from ", f.config.FileName)
	return nil
}

func (f *FileLoader) Flush(store tap.AuthRegisterBackend) error {
	oldSet, err := ioutil.ReadFile(f.config.FileName)
	if err != nil {
		log.Error("[FILE LOADER] load failed! ", err)
		return err
	}

	ts := strconv.Itoa(int(time.Now().Unix()))
	bkFilename := "profiles_backup_" + ts + ".json"
	bkLocation := path.Join(config.ProfileDir, bkFilename)

	wErr := ioutil.WriteFile(bkLocation, oldSet, 0644)
	if wErr != nil {
		log.Error("[FILE LOADER] backup failed! ", wErr)
		return wErr
	}

	newSet := store.GetAll()
	asJson, encErr := json.Marshal(newSet)
	if encErr != nil {
		log.Error("[FILE LOADER] Encoding failed! ", encErr)
		return encErr
	}

	savePath := path.Join(config.ProfileDir, f.config.FileName)
	if ProfileFilename != nil {
		savePath = path.Join(config.ProfileDir, *ProfileFilename)
	}

	w2Err := ioutil.WriteFile(savePath, asJson, 0644)
	if wErr != nil {
		log.Error("[FILE LOADER] flush failed! ", w2Err)
		return w2Err
	}

	return nil

}
