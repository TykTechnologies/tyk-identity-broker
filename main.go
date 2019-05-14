package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"path"
	"strconv"

	"github.com/TykTechnologies/tyk-identity-broker/backends"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
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

var log = logger.Get()

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
	log.Info("Tyk Identity Broker ", VERSION)
	log.Info("Copyright Tyk Technologies Ltd 2019")

	confFile := flag.String("c", "tib.conf", "Path to the config file")
	ProfileFilename := flag.String("p", "./profiles.json", "Path to the profiles file")
	flag.Parse()

	loadConfig(*confFile, &config)
	initBackend(config.BackEnd.ProfileBackendSettings, config.BackEnd.IdentityBackendSettings)

	TykAPIHandler = config.TykAPISettings

	// In OIDC there are calls to the https://{IDP-DOMAIN}/.well-know/openid-configuration and other endpoints
	// We set the http client's Transport to do InsecureSkipVerify to avoid error in case the certificate
	// was signed by unknown authority, trusting the user to set up his profile with the correct .well-know URL.
	http.DefaultClient.Transport =
		&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SSLInsecureSkipVerify}}

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

	p.Handle("/health", http.HandlerFunc(HandleHealthCheck)).Methods("GET")

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
