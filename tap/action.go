/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

// An Action is a value that defines what a particular authentication profile will do, for example, create and
// log in a user to the dashboard, or to the portal. Alternatively, create a token or OAuth session
type Action string

const (
	// Pass through / redirect user-based actions
	GenerateOrLoginDeveloperProfile Action = "GenerateOrLoginDeveloperProfile" // Portal
	GenerateOrLoginUserProfile      Action = "GenerateOrLoginUserProfile"      // Dashboard
	GenerateOAuthTokenForClient     Action = "GenerateOAuthTokenForClient"     // OAuth token flow

	// Direct or redirect
	GenerateTemporaryAuthToken    Action = "GenerateTemporaryAuthToken"  // Tyk Access Token
	GenerateOAuthTokenForPassword Action = "GenerateOAuthTokenForClient" // OAuth PW flow
)
