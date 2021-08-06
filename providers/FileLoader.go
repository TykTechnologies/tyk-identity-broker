package providers

import (
	"io/ioutil"
	"strings"
)

type FileLoader struct{}
var FileLoaderLogTag = "CERT FILE LOADER"
var FileLoaderLogger = log.WithField("prefix", SAMLLogTag)

func (f FileLoader) GetKey(key string) (string, error) {
	id := strings.Trim(key, "raw-")
	rawCert, err := ioutil.ReadFile(id)
	if err != nil {
		FileLoaderLogger.Errorf("Could not read cert file: %v", err.Error())
	}
	return string(rawCert), err
}

func (f FileLoader) SetKey(string, string, int64) error {
	panic("implement me")
}

func (f FileLoader) GetKeys(string) []string {
	panic("implement me")
}

func (f FileLoader) DeleteKey(string) bool {
	panic("implement me")
}

func (f FileLoader) DeleteScanMatch(string) bool {
	panic("implement me")
}

func (f FileLoader) GetListRange(string, int64, int64) ([]string, error) {
	panic("implement me")
}

func (f FileLoader) RemoveFromList(string, string) error {
	panic("implement me")
}

func (f FileLoader) AppendToSet(string, string) {
	panic("implement me")
}

func (f FileLoader) Exists(string) (bool, error) {
	panic("implement me")
}
