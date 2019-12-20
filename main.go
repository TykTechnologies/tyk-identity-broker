package main

import (
	"crypto/tls"
	"flag"
	"github.com/TykTechnologies/tyk-identity-broker/Initializer"
	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	"github.com/TykTechnologies/tyk-identity-broker/data_loader"
	"net/http"
	"strconv"

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
var config configuration.Configuration

// TykAPIHandler is a global API handler for Tyk, wraps the tyk APi in Go functions
var TykAPIHandler tyk.TykAPI

var GlobalDataLoader data_loader.DataLoader

var log = logger.Get()
var mainLogger = log.WithField("prefix", "MAIN")
var ProfileFilename *string


func init() {
	mainLogger.Info("Tyk Identity Broker ", VERSION)
	mainLogger.Info("Copyright Tyk Technologies Ltd 2019")

	confFile := flag.String("c", "tib.conf", "Path to the config file")
	ProfileFilename := flag.String("p", "./profiles.json", "Path to the profiles file")
	flag.Parse()

	configuration.LoadConfig(*confFile, &config)
	AuthConfigStore, IdentityKeyStore = Initializer.InitBackend(config.BackEnd.ProfileBackendSettings, config.BackEnd.IdentityBackendSettings)

	TykAPIHandler = config.TykAPISettings

	// In OIDC there are calls to the https://{IDP-DOMAIN}/.well-know/openid-configuration and other endpoints
	// We set the http client's Transport to do InsecureSkipVerify to avoid error in case the certificate
	// was signed by unknown authority, trusting the user to set up his profile with the correct .well-know URL.
	http.DefaultClient.Transport =
		&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SSLInsecureSkipVerify}}

	var err error
	GlobalDataLoader, err = data_loader.CreateDataLoader(config, ProfileFilename)
	if err != nil {
		return
	}
	err = GlobalDataLoader.LoadIntoStore(AuthConfigStore)
	if err != nil {
		mainLogger.Errorf("loading into store ",err)
		return
	}

	tothic.TothErrorHandler = HandleError
}

func main() {
	p := mux.NewRouter()
	p.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(HandleAuthCallback))
	p.Handle("/auth/{id}/{provider}", http.HandlerFunc(HandleAuth))

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
		mainLogger.Info("Broker Listening on SSL:", listenPort)
		err := http.ListenAndServeTLS(":"+listenPort, config.HttpServerOptions.CertFile, config.HttpServerOptions.KeyFile, p)
		if err != nil {
			mainLogger.Fatal("ListenAndServe: ", err)
		}
	} else {
		mainLogger.Info("Broker Listening on :", listenPort)
		http.ListenAndServe(":"+listenPort, p)
	}

}
