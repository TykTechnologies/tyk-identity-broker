/* package toth is a clone of goth, but modified for multi-tenant usage instead of using
globals everywhere */
package toth

import (
	"fmt"
	"github.com/markbates/goth"
)

// TothInstance wraps a goth configuration
type TothInstance struct {
	providers goth.Providers
}

// Init just creates the basic configuration objects
func (t *TothInstance) Init() {
	t.providers = goth.Providers{}
}

// UseProviders sets a list of available providers for use with Goth.
func (t *TothInstance) UseProviders(viders ...goth.Provider) {
	for _, provider := range viders {
		t.providers[provider.Name()] = provider
	}
}

// GetProviders returns a list of all the providers currently in use.
func (t *TothInstance) GetProviders() goth.Providers {
	return t.providers
}

// GetProvider returns a previously created provider. If Goth has not
// been told to use the named provider it will return an error.
func (t *TothInstance) GetProvider(name string) (goth.Provider, error) {
	provider := t.providers[name]
	if provider == nil {
		return nil, fmt.Errorf("no provider for %s exists", name)
	}
	return provider, nil
}

// ClearProviders will remove all providers currently in use.
// This is useful, mostly, for testing purposes.
func (t *TothInstance) ClearProviders() {
	t.providers = goth.Providers{}
}
