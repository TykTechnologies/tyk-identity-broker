package tap

type IdentityHandler interface {
	CreateIdentity(interface{}) (string, error)
	LoginIdentity(string, string) error
}

type DummyIdentityHandler struct{}

func (d DummyIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	return "", nil
}

func (d DummyIdentityHandler) LoginIdentity(user string, pass string) error {
	return nil
}
