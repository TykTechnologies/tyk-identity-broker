/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

// Profile is the configuration objct for an authentication session, it
// combines an Action (what to do with the identity once confirmed, this is
// delegated to an IdentityHandler) with a Provider (such as Social / GPlus)
type Profile struct {
	ID                    string
	OrgID                 string
	ActionType            Action
	MatchedPolicyID       string
	Type                  ProviderType
	ProviderName          string
	CustomEmailField      string
	CustomUserIDField     string
	ProviderConfig        interface{}
	IdentityHandlerConfig interface{}
	ProviderConstraints   ProfileConstraint
	ReturnURL             string
	DefaultUserGroupID    string
	CustomUserGroupField  string
	UserGroupMapping      map[string]string
}

// Certain providers can have constraints, this object sets out those constraints. E.g. Domain: "tyk.io" will limit
// social logins to only those with a tyk.io domain name
type ProfileConstraint struct {
	Domain string
	Group  string
}
