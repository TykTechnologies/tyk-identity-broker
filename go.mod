module github.com/TykTechnologies/tyk-identity-broker

go 1.13

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/TykTechnologies/tyk v1.9.2-0.20210112201019-11dba25d812b
	github.com/crewjam/saml v0.4.5
	github.com/go-ldap/ldap/v3 v3.2.3
	github.com/go-redis/redis/v8 v8.3.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/sessions v1.2.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/markbates/goth v1.64.2
	github.com/matryer/is v1.4.0
	github.com/mattermost/xml-roundtrip-validator v0.0.0-20201219040909-8fd2afad43d1 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.3-0.20191026113918-67a7fdcf741f
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.3
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

replace github.com/jeffail/tunny => github.com/Jeffail/tunny v0.0.0-20171107125207-452a8e97d6a3

replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20201012125356-562407e88c4f

exclude github.com/TykTechnologies/tyk/certs v0.0.1
