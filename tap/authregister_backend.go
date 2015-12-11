package tap

type AuthRegisterBackend interface {
	Init()
	SetKey(string, string)
	GetKey(string, string)
}
