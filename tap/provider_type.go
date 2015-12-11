package tap

type ProviderType string

const (
	PASSTHROUGH_PROVIDER ProviderType = "passthrough"
	REDIRECT_PROVIDER    ProviderType = "redirect"
)
