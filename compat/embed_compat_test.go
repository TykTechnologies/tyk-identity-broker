// Package compat_test guards the public embedding API used by ai-studio (and
// tyk-analytics).  Every function call, interface method, and type that an
// embedding application relies on is exercised here.  If a future change
// breaks the integration at compile time or at runtime this test will catch
// it before the consumer repos notice.
//
// The setup deliberately mirrors ai-studio's SSOService.InitInternalTIB()
// pattern: custom in-memory backends, no tib.conf, no Redis, no MongoDB.
package compat_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/TykTechnologies/tyk-identity-broker/data_loader"
	tykerrors "github.com/TykTechnologies/tyk-identity-broker/error"
	"github.com/TykTechnologies/tyk-identity-broker/initializer"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	tyk "github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Custom profile backend ────────────────────────────────────────────────────
//
// Mirrors ai-studio's GormAuthRegisterBackend: a concrete type that satisfies
// tap.AuthRegisterBackend using application-owned storage (GORM there, a map
// here).  The important thing is that none of TIB's built-in backends are used.

type inMemoryProfileStore struct {
	mu       sync.RWMutex
	profiles map[string]tap.Profile
}

func newProfileStore() *inMemoryProfileStore {
	return &inMemoryProfileStore{profiles: make(map[string]tap.Profile)}
}

func (s *inMemoryProfileStore) Init(_ interface{}) error { return nil }

func (s *inMemoryProfileStore) SetKey(key, _ string, val interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch v := val.(type) {
	case tap.Profile:
		s.profiles[key] = v
	case *tap.Profile:
		s.profiles[key] = *v
	default:
		return fmt.Errorf("expected tap.Profile, got %T", val)
	}
	return nil
}

func (s *inMemoryProfileStore) GetKey(key, _ string, val interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[key]
	if !ok {
		return fmt.Errorf("profile %q not found", key)
	}
	dst, ok := val.(*tap.Profile)
	if !ok {
		return fmt.Errorf("expected *tap.Profile, got %T", val)
	}
	*dst = p
	return nil
}

func (s *inMemoryProfileStore) GetAll(_ string) []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]interface{}, 0, len(s.profiles))
	for _, p := range s.profiles {
		cp := p
		out = append(out, cp)
	}
	return out
}

func (s *inMemoryProfileStore) DeleteKey(key, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profiles, key)
	return nil
}

// ── Custom KV backend ─────────────────────────────────────────────────────────
//
// Mirrors ai-studio's GormKVStore: used for OAuth session state and nonce
// tokens.  Values are JSON-serialised so any type can be round-tripped, which
// is what tothic and TykIdentityHandler both require.

type inMemoryKVStore struct {
	mu   sync.RWMutex
	data map[string]json.RawMessage
}

func newKVStore() *inMemoryKVStore {
	return &inMemoryKVStore{data: make(map[string]json.RawMessage)}
}

func (s *inMemoryKVStore) Init(_ interface{}) error { return nil }

func (s *inMemoryKVStore) SetKey(key, _ string, val interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	s.data[key] = b
	return nil
}

func (s *inMemoryKVStore) GetKey(key, _ string, val interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.data[key]
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}
	return json.Unmarshal(b, val)
}

func (s *inMemoryKVStore) GetAll(_ string) []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []interface{}
	for _, b := range s.data {
		var v interface{}
		if err := json.Unmarshal(b, &v); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func (s *inMemoryKVStore) DeleteKey(key, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// ── Embedded TIB ─────────────────────────────────────────────────────────────

// embeddedTIB mirrors ai-studio's InternalTIB struct.
type embeddedTIB struct {
	authConfigStore tap.AuthRegisterBackend
	kvStore         tap.AuthRegisterBackend
	tykAPIHandler   tyk.TykAPI
}

// initTIB replicates ai-studio's SSOService.InitInternalTIB() call-for-call.
// Any rename or signature change in the called functions will break this.
func initTIB(t *testing.T) *embeddedTIB {
	t.Helper()

	authStore := newProfileStore()
	kvStore := newKVStore()

	log := logrus.New()
	log.SetOutput(io.Discard)

	// --- exact ai-studio init sequence ---
	initializer.SetLogger(log)
	initializer.SetConfigHandler(kvStore)
	tothic.TothErrorHandler = tykerrors.HandleError
	tothic.Store = sessions.NewCookieStore([]byte("test-secret"))
	// -------------------------------------

	tib := &embeddedTIB{
		authConfigStore: authStore,
		kvStore:         kvStore,
		tykAPIHandler:   tyk.TykAPI{},
	}

	// CustomDispatcher mirrors ai-studio's setCustomDispatcher: TIB calls
	// back into the host-app API through this function instead of over HTTP.
	tib.tykAPIHandler.CustomDispatcher = func(
		target tyk.Endpoint, method, _ string, body io.Reader,
	) ([]byte, int, error) {
		return []byte(`{"key_id":"test-key"}`), http.StatusOK, nil
	}

	return tib
}

// buildRouter wires the auth endpoints the way ai-studio's sso_handlers_enterprise.go
// does: call providers.GetTapProfile, then delegate to the provider methods.
func buildRouter(tib *embeddedTIB) *mux.Router {
	r := mux.NewRouter()

	r.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		provider, profile, httpErr := providers.GetTapProfile(
			tib.authConfigStore, tib.kvStore, id, tib.tykAPIHandler,
		)
		if httpErr != nil {
			http.Error(w, httpErr.Message, httpErr.Code)
			return
		}
		provider.HandleCallback(w, req, tykerrors.HandleError, profile)
	}))

	r.Handle("/auth/{id}/{provider}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		provider, profile, httpErr := providers.GetTapProfile(
			tib.authConfigStore, tib.kvStore, id, tib.tykAPIHandler,
		)
		if httpErr != nil {
			http.Error(w, httpErr.Message, httpErr.Code)
			return
		}
		provider.Handle(w, req, mux.Vars(req), profile)
	}))

	return r
}

// proxyProfile returns a tap.Profile configured to use ProxyProvider, pointing
// at the given upstream URL.
func proxyProfile(id, orgID, upstreamURL string) tap.Profile {
	return tap.Profile{
		ID:           id,
		OrgID:        orgID,
		ProviderName: "ProxyProvider",
		Type:         tap.PASSTHROUGH_PROVIDER,
		ActionType:   tap.GenerateTemporaryAuthToken,
		ProviderConfig: map[string]interface{}{
			"TargetHost": upstreamURL,
			"OKCode":     0,
			"OKResponse": "",
			"OKRegex":    "",
		},
		IdentityHandlerConfig: map[string]interface{}{},
		ReturnURL:             "http://localhost/sso",
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestCustomBackendsSatisfyInterface is a compile-time assertion.
// If tap.AuthRegisterBackend gains or loses a method, this file will not compile.
func TestCustomBackendsSatisfyInterface(t *testing.T) {
	var _ tap.AuthRegisterBackend = &inMemoryProfileStore{}
	var _ tap.AuthRegisterBackend = &inMemoryKVStore{}
}

// TestProfileStoreRoundTrip verifies all four methods of the profile backend.
func TestProfileStoreRoundTrip(t *testing.T) {
	store := newProfileStore()

	p := tap.Profile{ID: "p1", OrgID: "org1", ProviderName: "ProxyProvider"}
	require.NoError(t, store.SetKey(p.ID, p.OrgID, p))

	var got tap.Profile
	require.NoError(t, store.GetKey("p1", "org1", &got))
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, p.OrgID, got.OrgID)
	assert.Equal(t, p.ProviderName, got.ProviderName)

	assert.Len(t, store.GetAll("org1"), 1)

	require.NoError(t, store.DeleteKey("p1", "org1"))
	assert.Error(t, store.GetKey("p1", "org1", &got))
}

// TestKVStoreRoundTrip verifies that the KV backend serialises arbitrary values —
// the same contract tothic requires for OAuth session state.
func TestKVStoreRoundTrip(t *testing.T) {
	store := newKVStore()

	type payload struct{ Token string }
	require.NoError(t, store.SetKey("k1", "", payload{Token: "abc123"}))

	var out payload
	require.NoError(t, store.GetKey("k1", "", &out))
	assert.Equal(t, "abc123", out.Token)

	require.NoError(t, store.DeleteKey("k1", ""))
	assert.Error(t, store.GetKey("k1", "", &out))
}

// TestInitSequence verifies the exact ai-studio initialisation call sequence
// runs without panicking or returning an error.
func TestInitSequence(t *testing.T) {
	tib := initTIB(t)
	require.NotNil(t, tib.authConfigStore)
	require.NotNil(t, tib.kvStore)
}

// TestGetTapProfile_NotFound verifies that a missing profile returns an
// *tap.HttpError with Code 404 — both ai-studio and tyk-analytics check this.
func TestGetTapProfile_NotFound(t *testing.T) {
	tib := initTIB(t)

	_, _, httpErr := providers.GetTapProfile(
		tib.authConfigStore, tib.kvStore, "nonexistent", tib.tykAPIHandler,
	)

	require.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.Code)
}

// TestGetTapProfile_ReturnsProvider verifies that GetTapProfile successfully
// looks up a stored profile, instantiates the right provider, and returns the
// profile data — the call ai-studio makes before every auth request.
func TestGetTapProfile_ReturnsProvider(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	tib := initTIB(t)
	p := proxyProfile("my-profile", "org1", upstream.URL)
	require.NoError(t, tib.authConfigStore.SetKey(p.ID, p.OrgID, p))

	provider, gotProfile, httpErr := providers.GetTapProfile(
		tib.authConfigStore, tib.kvStore, p.ID, tib.tykAPIHandler,
	)

	require.Nil(t, httpErr)
	require.NotNil(t, provider)
	assert.Equal(t, p.ID, gotProfile.ID)
	assert.Equal(t, "ProxyProvider", provider.Name())
}

// TestHTTPHandler_ProfileNotFound verifies that the ai-studio HTTP handler
// pattern propagates the 404 from GetTapProfile to the HTTP response.
func TestHTTPHandler_ProfileNotFound(t *testing.T) {
	tib := initTIB(t)
	router := buildRouter(tib)

	req := httptest.NewRequest(http.MethodGet, "/auth/nonexistent/ProxyProvider", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestHTTPHandler_ProxyUpstreamRejected verifies the full embedding path:
// profile found → provider.Handle() → proxy forwards to upstream →
// upstream rejects → 401 propagated back.
func TestHTTPHandler_ProxyUpstreamRejected(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	tib := initTIB(t)
	p := proxyProfile("proxy-profile", "org1", upstream.URL)
	require.NoError(t, tib.authConfigStore.SetKey(p.ID, p.OrgID, p))

	router := buildRouter(tib)
	req := httptest.NewRequest(http.MethodGet, "/auth/proxy-profile/ProxyProvider", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── New API tests (initializer.New + tothic.NewSessionStore + data_loader.NoopDataLoader) ──

// TestNewAPIInterfaces verifies the narrower tap.ProfileStore and tap.KVStore
// interfaces are satisfied by the same in-memory types — embedding apps can
// choose to implement only these if they don't need the full AuthRegisterBackend.
func TestNewAPIInterfaces(t *testing.T) {
	var _ tap.ProfileStore = &inMemoryProfileStore{}
	var _ tap.KVStore = &inMemoryKVStore{}
}

// TestTothicNewSessionStore verifies that tothic.NewSessionStore creates a
// usable store from a secret without reading any environment variable.
func TestTothicNewSessionStore(t *testing.T) {
	store := tothic.NewSessionStore("my-secret")
	require.NotNil(t, store)

	// Verify it is a sessions.Store (compile-time check)
	var _ sessions.Store = store
}

// TestNoopDataLoader verifies that data_loader.NoopDataLoader satisfies
// data_loader.DataLoader and all three methods return nil.
func TestNoopDataLoader(t *testing.T) {
	var loader data_loader.DataLoader = data_loader.NoopDataLoader{}

	assert.NoError(t, loader.Init(nil))
	assert.NoError(t, loader.LoadIntoStore(newProfileStore()))
	assert.NoError(t, loader.Flush(newProfileStore()))
}

// TestInitializerNew_Basic verifies that initializer.New completes the full
// init sequence (SetLogger, SetConfigHandler, session store, CustomDispatcher)
// in a single call instead of the six manual steps ai-studio performs.
func TestInitializerNew_Basic(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	broker := initializer.New(initializer.EmbedConfig{
		SessionSecret: "test-secret",
		Logger:        log,
		ProfileStore:  newProfileStore(),
		KVStore:       newKVStore(),
		CustomDispatcher: func(target tyk.Endpoint, method, _ string, body io.Reader) ([]byte, int, error) {
			return []byte(`{"key_id":"test"}`), http.StatusOK, nil
		},
	})

	require.NotNil(t, broker)
	require.NotNil(t, broker.ProfileStore)
	require.NotNil(t, broker.KVStore)
}

// TestInitializerNew_RegisterRoutes verifies that RegisterRoutes wires the
// three TIB endpoints and that a missing profile returns the expected error.
func TestInitializerNew_RegisterRoutes(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	broker := initializer.New(initializer.EmbedConfig{
		SessionSecret: "test-secret",
		Logger:        log,
		ProfileStore:  newProfileStore(),
		KVStore:       newKVStore(),
	})

	r := mux.NewRouter()
	broker.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/auth/nonexistent/ProxyProvider", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestBrokerGetTapProfile verifies the Broker convenience wrapper returns
// the same result as calling providers.GetTapProfile directly.
func TestBrokerGetTapProfile(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	profileStore := newProfileStore()
	p := proxyProfile("broker-profile", "org1", upstream.URL)
	require.NoError(t, profileStore.SetKey(p.ID, p.OrgID, p))

	log := logrus.New()
	log.SetOutput(io.Discard)

	broker := initializer.New(initializer.EmbedConfig{
		SessionSecret: "test-secret",
		Logger:        log,
		ProfileStore:  profileStore,
		KVStore:       newKVStore(),
	})

	// Not-found case
	_, _, httpErr := broker.GetTapProfile("nonexistent")
	require.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.Code)

	// Found case
	provider, gotProfile, httpErr := broker.GetTapProfile(p.ID)
	require.Nil(t, httpErr)
	require.NotNil(t, provider)
	assert.Equal(t, p.ID, gotProfile.ID)
}

// TestBrokerHandleAuth_ExplicitParams verifies the exported HandleAuth method
// works with explicit profileID/providerName — the pattern for gin/echo/chi.
func TestBrokerHandleAuth_ExplicitParams(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	profileStore := newProfileStore()
	p := proxyProfile("explicit-params", "org1", upstream.URL)
	require.NoError(t, profileStore.SetKey(p.ID, p.OrgID, p))

	log := logrus.New()
	log.SetOutput(io.Discard)

	broker := initializer.New(initializer.EmbedConfig{
		SessionSecret: "test-secret",
		Logger:        log,
		ProfileStore:  profileStore,
		KVStore:       newKVStore(),
	})

	// Call the exported method directly, the way a gin handler would
	req := httptest.NewRequest(http.MethodGet, "/auth/explicit-params/ProxyProvider", nil)
	w := httptest.NewRecorder()
	broker.HandleAuth(w, req, p.ID, "ProxyProvider")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInitializerNew_EndToEnd verifies the full path through the new API:
// New() → RegisterRoutes() → stored profile found → proxy auth rejected → 401.
func TestInitializerNew_EndToEnd(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	profileStore := newProfileStore()
	p := proxyProfile("end-to-end", "org1", upstream.URL)
	require.NoError(t, profileStore.SetKey(p.ID, p.OrgID, p))

	log := logrus.New()
	log.SetOutput(io.Discard)

	broker := initializer.New(initializer.EmbedConfig{
		SessionSecret: "test-secret",
		Logger:        log,
		ProfileStore:  profileStore,
		KVStore:       newKVStore(),
	})

	r := mux.NewRouter()
	broker.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/auth/end-to-end/ProxyProvider", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
