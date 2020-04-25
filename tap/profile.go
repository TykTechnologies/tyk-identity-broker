/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

// Profile is the configuration objct for an authentication session, it
// combines an Action (what to do with the identity once confirmed, this is
// delegated to an IdentityHandler) with a Provider (such as Social / GPlus)
type Profile struct {
	ID                     string            `bson:"ID" json:"ID"`
	OrgID                  string            `bson:"OrgID" json:"OrgID"`
	ActionType             Action            `bson:"ActionType" json:"ActionType"`
	MatchedPolicyID        string            `bson:"MatchedPolicyID" json:"MatchedPolicyID"`
	Type                   ProviderType      `bson:"Type" json:"Type"`
	ProviderName           string            `bson:"ProviderName" json:"ProviderName"`
	CustomEmailField       string            `bson:"CustomEmailField" json:"CustomEmailField"`
	CustomUserIDField      string            `bson:"CustomUserIDField" json:"CustomUserIDField"`
	ProviderConfig         interface{}       `bson:"ProviderConfig" json:"ProviderConfig"`
	IdentityHandlerConfig  interface{}       `bson:"IdentityHandlerConfig" json:"IdentityHandlerConfig"`
	ProviderConstraints    ProfileConstraint `bson:"ProviderConstraints" json:"ProviderConstraints"`
	ReturnURL              string            `bson:"ReturnURL" json:"ReturnURL"`
	DefaultUserGroupID     string            `bson:"DefaultUserGroupID" json:"DefaultUserGroupID"`
	CustomUserGroupField   string            `bson:"CustomUserGroupField" json:"CustomUserGroupField"`
	UserGroupMapping       map[string]string `bson:"UserGroupMapping" json:"UserGroupMapping"`
	CustomPortalGroupField string            `bson:"CustomPortalGroupField" json:"CustomPortalGroupField"`
	PortalGroupMapping     map[string]string `bson:"PortalGroupMapping" json:"PortalGroupMapping"`
}

// Certain providers can have constraints, this object sets out those constraints. E.g. Domain: "tyk.io" will limit
// social logins to only those with a tyk.io domain name
type ProfileConstraint struct {
	Domain string
	Group  string
}
