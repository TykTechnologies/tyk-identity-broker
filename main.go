package main

import (
	"crypto/tls"
	"flag"
	"net"
	"net/http"
	"strconv"

	"github.com/TykTechnologies/tyk-identity-broker/backends"
	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	"github.com/TykTechnologies/tyk-identity-broker/data_loader"
	"github.com/TykTechnologies/tyk-identity-broker/initializer"

	errors "github.com/TykTechnologies/tyk-identity-broker/error"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/gorilla/mux"
)

// config is the system-wide configuration
var config configuration.Configuration

// TykAPIHandler is a global API handler for Tyk, wraps the tyk APi in Go functions
var TykAPIHandler tyk.TykAPI

var GlobalDataLoader data_loader.DataLoader

var log = logger.Get()
var mainLogger = log.WithField("prefix", "MAIN")
var ProfileFilename, confFile string

func init() {
	mainLogger.Info("Tyk Identity Broker ", Version)
	mainLogger.Info("Copyright Tyk Technologies Ltd 2020")

	flag.StringVar(&confFile, "conf", "tib.conf", "Path to the config file")
	flag.StringVar(&confFile, "c", "tib.conf", "Path to the config file")
	flag.StringVar(&ProfileFilename, "p", "./profiles.json", "Path to the profiles file")
	flag.Parse()

	configuration.LoadConfig(confFile, &config)
	initializer.InitBackend(config.BackEnd.ProfileBackendSettings, config.BackEnd.IdentityBackendSettings)

	configStore := &backends.RedisBackend{KeyPrefix: "tib-provider-config-"}
	configStore.Init(config.BackEnd.IdentityBackendSettings)
	initializer.SetConfigHandler(configStore)

	TykAPIHandler = config.TykAPISettings

	// In OIDC there are calls to the https://{IDP-DOMAIN}/.well-know/openid-configuration and other endpoints
	// We set the http client's Transport to do InsecureSkipVerify to avoid error in case the certificate
	// was signed by unknown authority, trusting the user to set up his profile with the correct .well-know URL.
	http.DefaultClient.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SSLInsecureSkipVerify,
		},
	}
	var err error
	GlobalDataLoader, err = data_loader.CreateDataLoader(config, ProfileFilename)
	if err != nil {
		return
	}
	err = GlobalDataLoader.LoadIntoStore(initializer.AuthConfigStore)
	if err != nil {
		mainLogger.Errorf("loading into store %v", err)
		return
	}

	tothic.TothErrorHandler = errors.HandleError
	tothic.SetupSessionStore()
}

func main() {
	p := mux.NewRouter()
	p.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(HandleAuthCallback))
	p.Handle("/auth/{id}/{provider}", http.HandlerFunc(HandleAuth))
	p.Handle("/auth/{id}/saml/metadata", http.HandlerFunc(HandleMetadata))

	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleGetProfile))).Methods("GET")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleAddProfile))).Methods("POST")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleUpdateProfile))).Methods("PUT")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleDeleteProfile))).Methods("DELETE")

	p.Handle("/api/profiles", IsAuthenticated(http.HandlerFunc(HandleGetProfileList))).Methods("GET")

	p.Handle("/health", http.HandlerFunc(HandleHealthCheck)).Methods("GET")

	listenPort := 3010
	if config.Port != 0 {
		listenPort = config.Port
	}

	var tibServer net.Listener
	if config.HttpServerOptions.UseSSL {
		mainLogger.Info("--> Using SSL (https) for TIB")
		cert, err := tls.LoadX509KeyPair(config.HttpServerOptions.CertFile, config.HttpServerOptions.KeyFile)

		if err != nil {
			mainLogger.WithError(err).Error("loading cert file")
			return
		}

		cfg := tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: config.HttpServerOptions.SSLInsecureSkipVerify,
		}
		tibServer = createListener(listenPort, &cfg)
	} else {
		mainLogger.Info("--> Standard listener (http) for TIB")
		tibServer = createListener(listenPort, nil)
	}
	_ = http.Serve(tibServer, p)

}

func createListener(port int, tlsConfig *tls.Config) (listener net.Listener) {
	var err error
	addr := ":" + strconv.Itoa(port)

	if tlsConfig != nil {
		listener, err = tls.Listen("tcp", addr, tlsConfig)

		// to consume Dash api, then we skip as well the verification in the client side
		tr := &http.Transport{TLSClientConfig: tlsConfig}
		c := &http.Client{Transport: tr}
		tyk.SetHttpClient(c)
	} else {
		listener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		log.Panic("Server creation failed! ", err)
	}

	return
}
