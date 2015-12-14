package tap

type AuthRegisterBackend interface {
	Init()
	SetKey(string, interface{}) error
	GetKey(string, interface{}) error
}
