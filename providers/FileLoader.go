package providers

import (
	"io/ioutil"
	"strings"
)

type FileLoader struct{}

var FileLoaderLogTag = "CERT FILE LOADER"
var FileLoaderLogger = log.WithField("prefix", FileLoaderLogTag)

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

func (f FileLoader) AddToSet(string, string) {
	panic("implement me")
}

func (f FileLoader) AddToSortedSet(string, string, float64) {
	panic("implement me")
}

func (f FileLoader) Connect() bool {
	panic("implement me")
}

func (f FileLoader) Decrement(string) {
	panic("implement me")
}

func (f FileLoader) DeleteAllKeys() bool {
	panic("implement me")
}

func (f FileLoader) DeleteKeys([]string) bool {
	panic("implement me")
}

func (f FileLoader) DeleteRawKey(string) bool {
	panic("implement me")
}

func (f FileLoader) GetAndDeleteSet(string) []interface{} {
	panic("implement me")
}

func (f FileLoader) GetExp(string) (int64, error) {
	panic("implement me")
}

func (f FileLoader) GetKeyPrefix() string {
	panic("implement me")
}

func (f FileLoader) GetKeysAndValues() map[string]string {
	panic("implement me")
}

func (f FileLoader) GetKeysAndValuesWithFilter(string) map[string]string {
	panic("implement me")
}

func (f FileLoader) GetMultiKey([]string) ([]string, error) {
	panic("implement me")
}

func (f FileLoader) GetRawKey(string) (string, error) {
	panic("implement me")
}

func (f FileLoader) GetRollingWindow(key string, per int64, pipeline bool) (int, []interface{}) {
	panic("implement me")
}

func (f FileLoader) GetSet(string) (map[string]string, error) {
	panic("implement me")
}

func (f FileLoader) GetSortedSetRange(string, string, string) ([]string, []float64, error) {
	panic("implement me")
}

func (f FileLoader) IncrememntWithExpire(string, int64) int64 {
	panic("implement me")
}

func (f FileLoader) RemoveFromSet(string, string) {
	panic("implement me")
}

func (f FileLoader) RemoveSortedSetRange(string, string, string) error {
	panic("implement me")
}

func (f FileLoader) SetExp(string, int64) error {
	panic("implement me")
}

func (f FileLoader) SetRawKey(string, string, int64) error {
	panic("implement me")
}

func (f FileLoader) SetRollingWindow(key string, per int64, val string, pipeline bool) (int, []interface{}) {
	panic("implement me")
}
