package main

type ProviderType string

const (
	PASSTHROUGH_PROVIDER ProviderType = "passthrough"
	REDIRECT_PROVIDER    ProviderType = "redirect"
)
