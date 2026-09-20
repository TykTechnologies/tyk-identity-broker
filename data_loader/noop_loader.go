package data_loader

import "github.com/TykTechnologies/tyk-identity-broker/tap"

// NoopDataLoader is a DataLoader that performs no load or flush operations.
// Use it when profiles are managed entirely through the management API or the
// host application's own storage — no file or MongoDB source is needed.
// This is the pattern used by tyk-analytics (DumbLoader) and ai-studio.
type NoopDataLoader struct{}

func (NoopDataLoader) Init(_ interface{}) error                      { return nil }
func (NoopDataLoader) LoadIntoStore(_ tap.AuthRegisterBackend) error { return nil }
func (NoopDataLoader) Flush(_ tap.AuthRegisterBackend) error         { return nil }
