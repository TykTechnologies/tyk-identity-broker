package tap

type IdentityHandler interface {
	CreateIdentity(interface{}) (string, error)
	LoginIdentity(string, string) error
}
