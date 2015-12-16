package tap

type Profile struct {
	ID                  string
	OrgID               string
	ActionType          Action
	MatchedPolicyID     string
	Type                ProviderType
	ProviderName        string
	ProviderConfig      interface{}
	ProviderConstraints ProfileConstraint
	ReturnURL           string
}

type ProfileConstraint struct {
	Domain string
	Group  string
}
