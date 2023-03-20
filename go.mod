module github.com/TykTechnologies/tyk-identity-broker

go 1.16

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/TykTechnologies/storage v0.0.0-20230308174156-ed14b745c68b
	github.com/TykTechnologies/tyk v1.9.2-0.20211217130848-b04d51712be7
	github.com/crewjam/saml v0.4.12
	github.com/go-ldap/ldap/v3 v3.2.3
	github.com/go-redis/redis/v8 v8.3.1
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/sessions v1.2.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/markbates/goth v1.64.2
	github.com/matryer/is v1.4.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/sirupsen/logrus v1.4.3-0.20191026113918-67a7fdcf741f
	github.com/stretchr/testify v1.8.1
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.7
)

replace github.com/jeffail/tunny => github.com/Jeffail/tunny v0.0.0-20171107125207-452a8e97d6a3

replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20211213120648-56cd4003725b

replace gorm.io/gorm => github.com/TykTechnologies/gorm v1.20.7-0.20210409171139-b5c340f85ed0

exclude github.com/TykTechnologies/tyk/certs v0.0.1
