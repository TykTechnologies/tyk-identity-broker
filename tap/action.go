/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

// An Action is a alue that defines what a particular authentication profile will do, for example, create and
// log in a user to the dashboard, or to the portal. Alternatively, create a token or OAuth session
type Action string

const (
	GenerateOrLoginDeveloperProfile Action = "GenerateOrLoginDeveloperProfile" // Portal
	GenerateOrLoginUserProfile      Action = "GenerateOrLoginUserProfile"      // Dashboard
)
