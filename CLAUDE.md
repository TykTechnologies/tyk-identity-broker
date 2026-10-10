# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run tests (mirrors CI)
./bin/ci-tests.sh file           # file-based storage backend
./bin/ci-tests.sh mongo-mgo      # mgo driver
./bin/ci-tests.sh mongo-official  # official mongo driver

# Run a single test
go test -v -run TestName ./...
go test -v -run TestName ./providers/

# Run with race detector (as CI does)
go test -race -timeout 10m ./...

# Lint (CI uses golangci-lint v2.5.0; report written to golangci-lint-report.json)
golangci-lint run

# Format / vet
go fmt ./...
go vet ./...
goimports -w .

# Run locally (requires tib.conf and profiles.json)
go run main.go -c tib.conf -p ./
```

**CI pipeline order:** `go vet` → `gofmt -s` → `goimports` → `go test -race` per package.

**Environment variables for tests:**
- `TYK_IB_STORAGE_STORAGETYPE`: `file` | `mongo` (default: `file`)
- `TYK_IB_STORAGE_MONGOCONF_MONGOURL`: MongoDB connection string
- `TYK_IB_STORAGE_MONGOCONF_DRIVER`: `mgo` | `mongo-go`
- `TYK_IB_PROFILEDIR`: path searched for `profiles.json`
- All `tib.conf` fields can be overridden with `TYK_IB_` prefix (via `envconfig`). Set `TYK_IB_OMITCONFIGFILE=true` to ignore the config file entirely and rely only on env vars.

**Testing caveat — `package main`:** Tests in `api_test.go` live in `package main` alongside global variables (`AuthConfigStore`, `config`, etc.). Running `go test .` from the root requires a valid `tib.conf` because `init()` calls `configuration.LoadConfig`. Use `-c tib.conf` or set `TYK_IB_OMITCONFIGFILE=true` and supply all settings via env vars.

## Architecture

Tyk Identity Broker (TIB) is a delegated authentication proxy: it authenticates users against external identity providers (OAuth2/Social, LDAP/AD, SAML, proxy) and then takes a configured action within Tyk infrastructure (generate API token, create/login dashboard user, create/login portal developer).

### Request Flow

```
GET /auth/{profileID}/{provider}
  → HandleAuth() (http_handlers.go)
  → providers.GetTapProfile() loads Profile from AuthConfigStore (in-memory)
  → GetTAProvider() instantiates TAProvider + IdentityHandler via factory switch
  → provider.Handle() initiates auth (redirect or passthrough)

GET /auth/{profileID}/{provider}/callback
  → HandleAuthCallback()
  → provider.HandleCallback() completes auth
  → IdentityHandler.Handle() executes post-auth Action
  → redirect / JSON response
```

### Core Interfaces (`tap/`)

Every component plugs into three interfaces:

```go
TAProvider          // auth providers: Init, Handle, HandleCallback, HandleMetadata
IdentityHandler     // post-auth actions: Init, Handle
AuthRegisterBackend // storage: GetKey, SetKey, DeleteKey, GetAll
```

### Providers (`providers/`)

| Provider | Type | Description |
|---|---|---|
| Social | redirect | OAuth2/OIDC via `markbates/goth` (Google, GitHub, etc.) |
| ADProvider | passthrough | Active Directory / LDAP |
| SAMLProvider | redirect | SAML 2.0; exposes metadata at `/auth/{id}/saml/metadata` |
| ProxyProvider | passthrough | Delegates to upstream HTTP endpoint |

To add a new provider: implement `tap.TAProvider`, add a `case` in `GetTAProvider()` in `providers/tapProvider.go`, and register a constant in `constants/constants.go`.

### Identity Handlers (`tap/identity-handlers/`)

`TykIdentityHandler` handles all production actions by calling into Tyk Gateway or Dashboard APIs:

| Action | Effect |
|---|---|
| `GenerateOrLoginDeveloperProfile` | Create/login portal developer |
| `GenerateOrLoginUserProfile` | Create/login dashboard user (SSO via nonce) |
| `GenerateOAuthTokenForClient` | OAuth token → redirect with `#token=` fragment |
| `GenerateTemporaryAuthToken` | Direct Tyk API key response |

To add a new action: add a constant in `tap/action.go` and a `case` in `getIdentityHandler()` in `providers/tapProvider.go`.

### Backends (`backends/`)

- **In-memory** (`InMemoryBackend`): profile store; populated at startup from a DataLoader
- **Redis** (`RedisBackend`): identity/token cache shared across TIB instances (enables horizontal scaling); also stores OAuth session state (key prefix `tib-provider-config-`)
- **MongoDB** (`MongoBackend`): optional persistent profile storage loaded at boot

`AuthConfigStore` (in-memory) and `IdentityKeyStore` (Redis) are the two live backend instances, wired up in `initializer.InitBackend()`.

### Data Loaders (`data_loader/`)

`DataLoader` determines how profiles are initially read into `AuthConfigStore`. Selected at startup by `config.Storage.StorageType`:
- `file` (default) — reads `profiles.json` (or `ProfileDir`)
- `mongo` — reads from MongoDB using `mgo` or `mongo-go` driver

Profiles can be mutated at runtime via the management API; each mutating handler calls `GlobalDataLoader.Flush()` to persist changes back to the backing store.

### `tothic` Package

A multi-tenant wrapper around `markbates/goth`/`gothic`. Manages OAuth session state by persisting path params (profile ID + provider) to Redis instead of HTTP session cookies, so redirects can be matched back to the originating profile across stateless instances. `tothic.SetupSessionStore()` initialises the Gorilla cookie store using a key from env.

### Key Conventions

- **Multi-tenancy**: every Profile has an `OrgID`; Redis keys are `{OrgID}-{ProfileID}`
- **ProviderConfig marshaling**: `ProviderConfig` is `interface{}`; it is round-tripped through JSON bytes before being passed to a provider via `hackProviderConf()` in `providers/tapProvider.go`
- **API auth**: management endpoints (`/api/profiles/...`) require `Authorization: {config.Secret}` header; checked by `IsAuthenticated` middleware in `api.go`
- **Nonce-based SSO**: dashboard/portal logins use one-time tokens stored in Redis via the `tothic` package
- **SSL**: global HTTP client uses `SSLInsecureSkipVerify` for OIDC metadata discovery (controlled by `tib.conf`)
- **Dynamic profile reload**: profiles can be added/updated/deleted at runtime via the API; changes call `Flush()` to persist back to storage

### Embedding API (`initializer/`)

TIB is embedded inside tyk-analytics and ai-studio. The `initializer` package is the stable surface for that embedding.

**New high-level API (preferred for new embedders):**

```go
broker := initializer.New(initializer.EmbedConfig{
    SessionSecret:    cfg.APISecret,           // replaces TYK_IB_SESSION_SECRET env var
    Logger:           myLogger,                // optional logrus.Logger
    ProfileStore:     myDB,                    // tap.AuthRegisterBackend — profile lookup
    KVStore:          myKV,                    // tap.AuthRegisterBackend — OAuth state + nonces
    CustomDispatcher: func(target tyk.Endpoint, method, _ string, body io.Reader) ([]byte, int, error) {
        // route TIB→host API calls in-process (standard pattern in both ai-studio and tyk-analytics)
        return routeInternally(target, method, body)
    },
})

// gorilla/mux: registers /auth/{id}/{provider}, /auth/{id}/{provider}/callback, /auth/{id}/saml/metadata
broker.RegisterRoutes(router)

// gin/echo/chi: extract params yourself and call the exported methods directly
broker.HandleAuth(w, r, profileID, providerName)
broker.HandleCallback(w, r, profileID)
broker.HandleMetadata(w, r, profileID)

// convenience wrapper around providers.GetTapProfile
provider, profile, httpErr := broker.GetTapProfile(profileID)
```

Implement `tap.AuthRegisterBackend` for your storage engine. For read-only profile lookup `tap.ProfileStore` (only `GetKey`) and for session state `tap.KVStore` (`GetKey`/`SetKey`/`DeleteKey`) are the narrower interfaces — any `AuthRegisterBackend` implementation satisfies both automatically.

Use `data_loader.NoopDataLoader` when profiles are managed via the API or your own DB (not a file or MongoDB source).

Use `tothic.NewSessionStore(secret)` to create the session store from a config value instead of the `TYK_IB_SESSION_SECRET` env var.

**Lower-level API (used by tyk-analytics, still fully supported):**

- `InitBackend(profileConf, identityConf)` — wires up `InMemoryBackend` + `RedisBackend`
- `CreateBackendFromRedisConn(kv, prefix)` — creates a `RedisBackend` from an existing connection
- `CreateInMemoryBackend()` / `CreateMongoBackend(store)` — construct specific backends
- `SetLogger(logger)` — injects the host process's logger
- `SetCertManager(cm)` — injects the Dashboard's certificate manager into SAML/OIDC providers

**Backward compatibility:** All existing embedder call sites (`SetLogger`, `SetConfigHandler`, `tothic.TothErrorHandler`, `tothic.Store`, `providers.GetTapProfile`) remain unchanged. The new API is purely additive.

The `compat/` package contains integration tests that guard every function and interface an external embedder calls. Run `go test ./compat/` after any change to the embedding surface.

### Configuration

- `tib.conf` — main JSON config (`Secret`, `SSL`, `BackEnd`, `TykAPISettings`, `Storage`); see `tib_sample.conf` for a minimal example
- `profiles.json` — array of `tap.Profile` objects; loaded at startup, cached in memory
- All fields overridable via env vars with prefix `TYK_IB_` (e.g. `TYK_IB_SECRET`, `TYK_IB_PORT`)
