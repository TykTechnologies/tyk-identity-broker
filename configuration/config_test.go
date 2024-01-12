package configuration

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/TykTechnologies/storage/persistent"

	"github.com/matryer/is"
)

func TestOverrideConfigWithEnvVars(t *testing.T) {
	is := is.New(t)

	secret := "SECRET"
	port := 1234
	profileDir := "PROFILEDIR"

	is.NoErr(os.Setenv("TYK_IB_SECRET", secret))
	is.NoErr(os.Setenv("TYK_IB_PORT", strconv.Itoa(port)))
	is.NoErr(os.Setenv("TYK_IB_PROFILEDIR", profileDir))
	is.NoErr(os.Setenv("TYK_IB_SSLINSECURESKIPVERIFY", "true"))

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
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_MAXIDLE", strconv.Itoa(maxIdle)))
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_MAXACTIVE", strconv.Itoa(maxActive)))
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_DATABASE", strconv.Itoa(database)))
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_PASSWORD", password))
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_ENABLECLUSTER", "true"))
	is.NoErr(os.Setenv("TYK_IB_BACKEND_IDENTITYBACKENDSETTINGS_HOSTS", hostsStr))

	// TykAPISettings.GatewayConfig
	gwEndpoint := "http://dummyhost"
	gwPort := "7890"
	gwAdminSecret := "76543"
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_ENDPOINT", gwEndpoint))
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_PORT", gwPort))
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_GATEWAYCONFIG_ADMINSECRET", gwAdminSecret))

	// TykAPISettings.DashboardConfig
	dbEndpoint := "http://dummyhost2"
	dbPort := "9876"
	dbAdminSecret := "87654"
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_ENDPOINT", dbEndpoint))
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_PORT", dbPort))
	is.NoErr(os.Setenv("TYK_IB_TYKAPISETTINGS_DASHBOARDCONFIG_ADMINSECRET", dbAdminSecret))

	// HttpServerOptions
	certFile := "./certs/server.pem"
	keyFile := "./certs/key.pem"
	is.NoErr(os.Setenv("TYK_IB_HTTPSERVEROPTIONS_USESSL", "true"))
	is.NoErr(os.Setenv("TYK_IB_HTTPSERVEROPTIONS_CERTFILE", certFile))
	is.NoErr(os.Setenv("TYK_IB_HTTPSERVEROPTIONS_KEYFILE", keyFile))

	// Assertions
	var conf Configuration
	LoadConfig("testdata/tib_test.conf", &conf)

	is.Equal(secret, conf.Secret)
	is.Equal(port, conf.Port)
	is.Equal(profileDir, conf.ProfileDir)
	is.Equal(true, conf.SSLInsecureSkipVerify)

	is.Equal(maxIdle, conf.BackEnd.IdentityBackendSettings.MaxIdle)
	is.Equal(maxActive, conf.BackEnd.IdentityBackendSettings.MaxActive)
	is.Equal(database, conf.BackEnd.IdentityBackendSettings.Database)
	is.Equal(password, conf.BackEnd.IdentityBackendSettings.Password)
	is.Equal(true, conf.BackEnd.IdentityBackendSettings.EnableCluster)
	is.Equal(hosts, conf.BackEnd.IdentityBackendSettings.Hosts)

	is.Equal(gwEndpoint, conf.TykAPISettings.GatewayConfig.Endpoint)
	is.Equal(gwPort, conf.TykAPISettings.GatewayConfig.Port)
	is.Equal(gwAdminSecret, conf.TykAPISettings.GatewayConfig.AdminSecret)
	is.Equal(dbEndpoint, conf.TykAPISettings.DashboardConfig.Endpoint)
	is.Equal(dbPort, conf.TykAPISettings.DashboardConfig.Port)
	is.Equal(dbAdminSecret, conf.TykAPISettings.DashboardConfig.AdminSecret)

	is.Equal(true, conf.HttpServerOptions.UseSSL)
	is.Equal(certFile, conf.HttpServerOptions.CertFile)
	is.Equal(keyFile, conf.HttpServerOptions.KeyFile)
}

func TestGetMongoDriver(t *testing.T) {
	tests := []struct {
		name           string
		driverFromConf string
		expected       string
	}{
		{
			name:           "valid persistent.Mgo",
			driverFromConf: persistent.Mgo,
			expected:       persistent.Mgo,
		},
		{
			name:           "valid persistent.OfficialMongo",
			driverFromConf: persistent.OfficialMongo,
			expected:       persistent.OfficialMongo,
		},
		{
			name:           "invalid driverFromConf",
			driverFromConf: "invalidDriver",
			expected:       persistent.Mgo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMongoDriver(tt.driverFromConf); got != tt.expected {
				t.Errorf("GetMongoDriver() = %v, want %v", got, tt.expected)
			}
		})
	}
}
