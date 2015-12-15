package tap

type AuthRegisterBackend interface {
	Init(interface{})
	SetKey(string, interface{}) error
	GetKey(string, interface{}) error
}
