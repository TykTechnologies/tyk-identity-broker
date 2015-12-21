/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

// AuthRegisterBackend is an interface to provide storage for profiles loaded into TAP
type AuthRegisterBackend interface {
	Init(interface{})
	SetKey(string, interface{}) error
	GetKey(string, interface{}) error
}
