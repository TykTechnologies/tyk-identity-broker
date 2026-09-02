package initializer

import (
	"io"
	"net/http"

	tykerrors "github.com/TykTechnologies/tyk-identity-broker/error"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	tyk "github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// EmbedConfig holds all settings needed to embed TIB into a host application.
// Construct one and pass it to New — no tib.conf, no env vars required.
//
// Minimal example (mirrors ai-studio's SSOService.InitInternalTIB):
//
//	broker := initializer.New(initializer.EmbedConfig{
//	    SessionSecret: cfg.APISecret,
//	    Logger:        myLogger,
//	    ProfileStore:  myGormBackend,
//	    KVStore:       myGormKVStore,
//	    CustomDispatcher: func(target tyk.Endpoint, method, _ string, body io.Reader) ([]byte, int, error) {
//	        // route TIB → host-app API calls through your own router
//	        return routeInternally(target, method, body)
//	    },
//	})
//	broker.RegisterRoutes(router)
type EmbedConfig struct {
	// SessionSecret signs the Gorilla cookie store used for OAuth session state.
	// Replaces the TYK_IB_SESSION_SECRET env var that SetupSessionStore reads.
	// Falls back to SetupSessionStore (env var) when empty.
	SessionSecret string

	// Logger replaces TIB's default logrus logger. Optional — omit to keep
	// the default logger.
	Logger *logrus.Logger

	// ProfileStore is the backend used to look up authentication profiles by ID.
	// Back it with your application's database; only GetKey is called on this
	// path (profile writes come through the management API, not auth flows).
	ProfileStore tap.AuthRegisterBackend

	// KVStore is the backend for OAuth session state and nonce tokens.
	// tothic and TykIdentityHandler both read, write, and delete ephemeral keys
	// here. A Redis-backed or GORM-backed implementation both work.
	KVStore tap.AuthRegisterBackend

	// CustomDispatcher intercepts every call TIB makes to Tyk Dashboard or
	// Gateway and routes it through the host application's own HTTP router
	// in-process — no outbound network hop. This is the standard pattern used
	// by both ai-studio and tyk-analytics.
	//
	// Signature matches tyk.TykAPI.CustomDispatcher exactly for easy migration.
	// When nil, TIB makes real HTTP calls using GatewayConfig / DashboardConfig.
	CustomDispatcher func(target tyk.Endpoint, method, usercode string, body io.Reader) ([]byte, int, error)

	// GatewayConfig and DashboardConfig are only consulted when CustomDispatcher
	// is nil (i.e. when TIB is running standalone, not embedded).
	GatewayConfig   tyk.EndpointConfig
	DashboardConfig tyk.EndpointConfig
}

// Broker is an initialised, embedded TIB instance.
// Use RegisterRoutes to attach its auth endpoints to your HTTP router.
type Broker struct {
	// ProfileStore is the backend used to look up authentication profiles.
	ProfileStore tap.AuthRegisterBackend
	// KVStore is the backend for OAuth session state and nonce tokens.
	KVStore    tap.AuthRegisterBackend
	apiHandler tyk.TykAPI
}

// New initialises TIB for in-process embedding. It calls SetLogger,
// SetConfigHandler, and configures the session store in one shot —
// the same sequence that ai-studio's SSOService.InitInternalTIB() and
// tyk-analytics's InitInternalTIB() perform across multiple manual calls.
func New(cfg EmbedConfig) *Broker {
	if cfg.Logger != nil {
		SetLogger(cfg.Logger)
	}

	SetConfigHandler(cfg.KVStore)

	tothic.TothErrorHandler = tykerrors.HandleError
	if cfg.SessionSecret != "" {
		tothic.Store = tothic.NewSessionStore(cfg.SessionSecret)
	} else {
		tothic.SetupSessionStore()
	}

	apiHandler := tyk.TykAPI{
		GatewayConfig:   cfg.GatewayConfig,
		DashboardConfig: cfg.DashboardConfig,
	}
	if cfg.CustomDispatcher != nil {
		apiHandler.CustomDispatcher = cfg.CustomDispatcher
	}

	return &Broker{
		ProfileStore: cfg.ProfileStore,
		KVStore:      cfg.KVStore,
		apiHandler:   apiHandler,
	}
}

// GetTapProfile looks up a profile by ID and instantiates its provider.
// It is a convenience wrapper around providers.GetTapProfile that binds the
// Broker's backends and API handler — the same operation that ai-studio's
// SSOService.GetTapProfile() performs manually.
func (b *Broker) GetTapProfile(id string) (tap.TAProvider, tap.Profile, *tap.HttpError) {
	return providers.GetTapProfile(b.ProfileStore, b.KVStore, id, b.apiHandler)
}

// RegisterRoutes wires TIB's three auth endpoints into r (gorilla/mux).
// Call this after New, before starting your HTTP server.
//
// Routes registered:
//
//	GET/POST /auth/{id}/{provider}           — initiates authentication
//	GET/POST /auth/{id}/{provider}/callback  — OAuth/OIDC callback
//	GET/POST /auth/{id}/saml/metadata        — SAML SP metadata
//
// For non-mux routers (gin, echo, chi) extract the path parameters yourself
// and call HandleAuth, HandleCallback, and HandleMetadata directly.
func (b *Broker) RegisterRoutes(r *mux.Router) {
	r.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(b.handleCallbackMux))
	r.Handle("/auth/{id}/{provider}", http.HandlerFunc(b.handleAuthMux))
	r.Handle("/auth/{id}/saml/metadata", http.HandlerFunc(b.handleMetadataMux))
}

// HandleAuth initiates authentication for the given profile and provider.
// Use this when integrating with routers other than gorilla/mux (gin, echo,
// chi, etc.) — extract profileID and providerName from the URL yourself and
// call this directly.
func (b *Broker) HandleAuth(w http.ResponseWriter, r *http.Request, profileID, providerName string) {
	params := map[string]string{"id": profileID, "provider": providerName}
	provider, profile, httpErr := providers.GetTapProfile(b.ProfileStore, b.KVStore, profileID, b.apiHandler)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.Code)
		return
	}
	provider.Handle(w, r, params, profile)
}

// HandleCallback completes an OAuth/OIDC callback for the given profile.
// Use this with non-mux routers.
func (b *Broker) HandleCallback(w http.ResponseWriter, r *http.Request, profileID string) {
	provider, profile, httpErr := providers.GetTapProfile(b.ProfileStore, b.KVStore, profileID, b.apiHandler)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.Code)
		return
	}
	provider.HandleCallback(w, r, tykerrors.HandleError, profile)
}

// HandleMetadata serves SAML SP metadata for the given profile.
// Use this with non-mux routers.
func (b *Broker) HandleMetadata(w http.ResponseWriter, r *http.Request, profileID string) {
	provider, _, httpErr := providers.GetTapProfile(b.ProfileStore, b.KVStore, profileID, b.apiHandler)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.Code)
		return
	}
	provider.HandleMetadata(w, r)
}

// mux-specific wrappers: extract path params from gorilla/mux, then delegate.

func (b *Broker) handleAuthMux(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	b.HandleAuth(w, r, vars["id"], vars["provider"])
}

func (b *Broker) handleCallbackMux(w http.ResponseWriter, r *http.Request) {
	b.HandleCallback(w, r, mux.Vars(r)["id"])
}

func (b *Broker) handleMetadataMux(w http.ResponseWriter, r *http.Request) {
	b.HandleMetadata(w, r, mux.Vars(r)["id"])
}
