package data_loader

import "github.com/TykTechnologies/tyk-identity-broker/tap"

// DumbLoader does nothing, use for those cases where cache not needed so it's the same data store
// so call Flush and LoadIntoStore doesnt make sense
type DumbLoader struct{}

func (DumbLoader) Init(conf interface{}) error {
	return nil
}

func (DumbLoader) LoadIntoStore(tap.AuthRegisterBackend) error {
	return nil
}

func (DumbLoader) Flush(tap.AuthRegisterBackend) error {
	return nil
}
