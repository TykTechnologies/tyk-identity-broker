package data_loader

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"io/ioutil"
	"path"
	"strconv"
	"time"
)

// FileLoaderConf is the configuration struct for a FileLoader, takes a filename as main init
type FileLoaderConf struct {
	FileName string
	ProfileDir string
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
	profiles := []tap.Profile{}

	thisSet, err := ioutil.ReadFile(f.config.FileName)
	if err != nil {
		dataLogger.WithFields(logrus.Fields{
			"filename": f.config.FileName,
			"error":    err,
		}).Error("Load failure")
		return err
	} else {
		jsErr := json.Unmarshal(thisSet, &profiles)
		if jsErr != nil {
			dataLogger.WithField("error", jsErr).Error("Couldn't unmarshal profile set")
			return err
		}
	}

	var loaded int
	for _, profile := range profiles {
		inputErr := store.SetKey(profile.ID, profile)
		if inputErr != nil {
			dataLogger.WithField("error", inputErr).Error("Couldn't encode configuration")
		} else {
			loaded += 1
		}
	}

	dataLogger.WithField("filename", f.config.FileName).Infof("Loaded %d profiles", loaded)
	return nil
}

func (f *FileLoader) Flush(store tap.AuthRegisterBackend) error {
	oldSet, err := ioutil.ReadFile(f.config.FileName)
	if err != nil {
		dataLogger.WithFields(logrus.Fields{
			"filename": f.config.FileName,
			"error":    err,
		}).Error("load failed!")
		return err
	}

	ts := strconv.Itoa(int(time.Now().Unix()))
	bkFilename := "profiles_backup_" + ts + ".json"
	bkLocation := path.Join(f.config.ProfileDir, bkFilename)

	wErr := ioutil.WriteFile(bkLocation, oldSet, 0644)
	if wErr != nil {
		dataLogger.WithFields(logrus.Fields{
			"bk_filename": bkFilename,
			"error":       err,
		}).Error("backup failed! ", wErr)
		return wErr
	}

	newSet := store.GetAll()
	asJson, encErr := json.Marshal(newSet)
	if encErr != nil {
		dataLogger.WithField("error", encErr).Error("Encoding failed!")
		return encErr
	}

	savePath := path.Join(f.config.ProfileDir, f.config.FileName)

	w2Err := ioutil.WriteFile(savePath, asJson, 0644)
	if wErr != nil {
		dataLogger.WithField("error", w2Err).Error("flush failed!")
		return w2Err
	}

	return nil

}
