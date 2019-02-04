package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestOverrideConfigWithEnvVars(t *testing.T) {
	secret := "SECRET"
	port := 1234
	profileDir := "PROFILEDIR"

	_ = os.Setenv("TYK_IB_SECRET", secret)
	_ = os.Setenv("TYK_IB_PORT", strconv.Itoa(port))
	_ = os.Setenv("TYK_IB_PROFILEDIR", profileDir)
	_ = os.Setenv("TYK_IB_SSLINSECURESKIPVERIFY", "true")

	// Backend
	maxIdle := 1020
	maxActive := 2020
	database := 1
	password := "PASSWORD"
	hosts := map[string]string{
		"dummyhost1": "1234",
		"dummyhost2": "5678",
	}
	var hostsStr string
	for key, value := range hosts {
		if hostsStr != "" {
			hostsStr += ","
		}
		hostsStr += fmt.Sprintf("%s:%s", key, value)
	}
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_MAXIDLE", strconv.Itoa(maxIdle))
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_MAXACTIVE", strconv.Itoa(maxActive))
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_DATABASE", strconv.Itoa(database))
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_PASSWORD", password)
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_ENABLECLUSTER", "true")
	_ = os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_HOSTS", hostsStr)

	// TykAPISettings.GatewayConfig
	gwEndpoint := "http://dummyhost"
	gwPort := "7890"
	gwAdminSecret := "76543"
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_ENDPOINT", gwEndpoint)
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_PORT", gwPort)
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_ADMINSECRET", gwAdminSecret)

	// TykAPISettings.DashboardConfig
	dbEndpoint := "http://dummyhost2"
	dbPort := "9876"
	dbAdminSecret := "87654"
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_ENDPOINT", dbEndpoint)
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_PORT", dbPort)
	_ = os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_ADMINSECRET", dbAdminSecret)

	// HttpServerOptions
	certFile := "./certs/server.pem"
	keyFile := "./certs/key.pem"
	_ = os.Setenv("TYK_IB_HTTPSERVEROPTIONS_USESSL", "true")
	_ = os.Setenv("TYK_IB_HTTPSERVEROPTIONS_CERTFILE", certFile)
	_ = os.Setenv("TYK_IB_HTTPSERVEROPTIONS_KEYFILE", keyFile)

	// Assertions
	var conf Configuration
	loadConfig("tib_sample.conf", &conf)

	assert(t, secret, conf.Secret)
	assert(t, port, conf.Port)
	assert(t, profileDir, conf.ProfileDir)
	assert(t, true, conf.SSLInsecureSkipVerify)

	assert(t, maxIdle, conf.BackEnd.IdentityBackendSettings.MaxIdle)
	assert(t, maxActive, conf.BackEnd.IdentityBackendSettings.MaxActive)
	assert(t, database, conf.BackEnd.IdentityBackendSettings.Database)
	assert(t, password, conf.BackEnd.IdentityBackendSettings.Password)
	assert(t, true, conf.BackEnd.IdentityBackendSettings.EnableCluster)
	assert(t, hosts, conf.BackEnd.IdentityBackendSettings.Hosts)

	assert(t, gwEndpoint, conf.TykAPISettings.GatewayConfig.Endpoint)
	assert(t, gwPort, conf.TykAPISettings.GatewayConfig.Port)
	assert(t, gwAdminSecret, conf.TykAPISettings.GatewayConfig.AdminSecret)
	assert(t, dbEndpoint, conf.TykAPISettings.DashboardConfig.Endpoint)
	assert(t, dbPort, conf.TykAPISettings.DashboardConfig.Port)
	assert(t, dbAdminSecret, conf.TykAPISettings.DashboardConfig.AdminSecret)

	assert(t, true, conf.HttpServerOptions.UseSSL)
	assert(t, certFile, conf.HttpServerOptions.CertFile)
	assert(t, keyFile, conf.HttpServerOptions.KeyFile)

}

func assert(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, actual %v", expected, actual)
	}
}
