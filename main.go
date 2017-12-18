package main

import (
	"flag"
	"net/http"
	"path"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/gorilla/mux"
)

// AuthConfigStore Is the back end we are storing our configuration files to
var AuthConfigStore tap.AuthRegisterBackend

// IdentityKeyStore keeps a record of identities tied to tokens (if needed)
var IdentityKeyStore tap.AuthRegisterBackend

//  config is the system-wide configuration
var config Configuration

// TykAPIHandler is a global API handler for Tyk, wraps the tyk APi in Go functions
var TykAPIHandler tyk.TykAPI

var GlobalDataLoader DataLoader

var log = logrus.New()

var ProfileFilename *string

// Get our bak end to use, new beack-ends must be registered here
func initBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{}) {

	AuthConfigStore = &backends.InMemoryBackend{}
	IdentityKeyStore = &backends.RedisBackend{KeyPrefix: "identity-cache."}

	log.Info("[MAIN] Initialising Profile Configuration Store")
	AuthConfigStore.Init(profileBackendConfiguration)
	log.Info("[MAIN] Initialising Identity Cache")
	IdentityKeyStore.Init(identityBackendConfiguration)
}

func init() {
	log.Level = logrus.DebugLevel

	log.Info("Tyk Identity Broker v0.2")
	log.Info("Copyright Martin Buhr 2016\n")

	confFile := flag.String("c", "tib.conf", "Path to the config file")
	ProfileFilename := flag.String("p", "./profiles.json", "Path to the profiles file")
	flag.Parse()

	loadConfig(*confFile, &config)
	initBackend(config.BackEnd.ProfileBackendSettings, config.BackEnd.IdentityBackendSettings)

	TykAPIHandler = config.TykAPISettings

	pDir := path.Join(config.ProfileDir, *ProfileFilename)
	loaderConf := FileLoaderConf{
		FileName: pDir,
	}

	GlobalDataLoader = &FileLoader{}
	GlobalDataLoader.Init(loaderConf)
	GlobalDataLoader.LoadIntoStore(AuthConfigStore)

	tothic.TothErrorHandler = HandleError
}

func main() {
	p := mux.NewRouter()
	p.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(HandleAuthCallback))
	p.Handle("/auth/{id}/{provider}", http.HandlerFunc(HandleAuth))

	p.Handle("/api/profiles/save", IsAuthenticated(http.HandlerFunc(HandleFlushProfileList))).Methods("POST")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleGetProfile))).Methods("GET")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleAddProfile))).Methods("POST")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleUpdateProfile))).Methods("PUT")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleDeleteProfile))).Methods("DELETE")

	p.Handle("/api/profiles", IsAuthenticated(http.HandlerFunc(HandleGetProfileList))).Methods("GET")

	listenPort := "3010"
	if config.Port != 0 {
		listenPort = strconv.Itoa(config.Port)
	}

	if config.HttpServerOptions.UseSSL {
		log.Info("[MAIN] Broker Listening on SSL:", listenPort)
		err := http.ListenAndServeTLS(":"+listenPort, config.HttpServerOptions.CertFile, config.HttpServerOptions.KeyFile, p)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		log.Info("[MAIN] Broker Listening on :", listenPort)
		http.ListenAndServe(":"+listenPort, p)
	}

}
