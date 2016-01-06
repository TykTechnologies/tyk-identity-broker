package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/lonelycode/tyk-auth-proxy/backends"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tothic"
	"github.com/lonelycode/tyk-auth-proxy/tyk-api"
	"net/http"
)

// AuthConfigStore Is the back end we are storing our configuration files to
var AuthConfigStore tap.AuthRegisterBackend

// IdentityKeyStore keeps a record of identities tied to tokens (if needed)
var IdentityKeyStore tap.AuthRegisterBackend

//  config is the system-wide configuration
var config Configuration

// TykAPIHandler is a global API handler for Tyk, wraps the tyk APi in Go functions
var TykAPIHandler tyk.TykAPI

var log = logrus.New()

// Get our bak end to use, new beack-ends must be registered here
func initBackend(name string, configuration interface{}) {
	found := false

	switch name {
	case "in_memory":
		AuthConfigStore = &backends.InMemoryBackend{}
		IdentityKeyStore = &backends.InMemoryBackend{}
		found = true
	}

	if !found {
		log.Warning("[MAIN] No backend set!")
		AuthConfigStore = &backends.InMemoryBackend{}
		IdentityKeyStore = &backends.InMemoryBackend{}

	}

	AuthConfigStore.Init(configuration)
	IdentityKeyStore.Init(configuration)
}

func init() {
	log.Info("Tyk Identity Broker v0.1")
	log.Info("Copyright Martin Buhr 2016\n")

	loadConfig("tib.conf", &config)
	initBackend(config.BackEnd.Name, config.BackEnd.BackendSettings)

	TykAPIHandler = config.TykAPISettings

	// --- Testing

	loaderConf := FileLoaderConf{
		FileName: "./test_apps.json",
	}

	loader := FileLoader{}
	loader.Init(loaderConf)
	loader.LoadIntoStore(AuthConfigStore)

	// --- End test

	tothic.TothErrorHandler = HandleError
}

func main() {
	p := mux.NewRouter()
	p.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(HandleAuthCallback))
	p.Handle("/auth/{id}/{provider}", http.HandlerFunc(HandleAuth))

	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleGetProfile))).Methods("GET")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleCreateProfile))).Methods("POST")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleUpdateProfile))).Methods("PUT")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleDeleteProfile))).Methods("DELETE")

	p.Handle("/api/profiles", IsAuthenticated(http.HandlerFunc(HandleGetProfileList))).Methods("GET")

	p.Handle("/api/profiles", IsAuthenticated(http.HandlerFunc(HandleGenerateProfile))).Methods("POST")

	log.Info("[MAIN] --> Server Listening...")
	http.ListenAndServe(":3010", p)
}
