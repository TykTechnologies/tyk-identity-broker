package identityHandlers

import (
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

type DummyIdentityHandler struct{}

func (d DummyIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	log.Info("Creating identity for: ", i)
	return "", nil
}

func (d DummyIdentityHandler) LoginIdentity(user string, pass string) error {
	log.Info("Logging in identity: ", user)
	return nil
}
