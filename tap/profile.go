package tap

type Profile struct {
	ID              string
	OrgID           string
	ActionType      Action
	MatchedPolicyID string
	Type            ProviderType
	ProviderName    string
	ProviderConfig  string
}
