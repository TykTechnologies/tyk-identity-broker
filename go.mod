module github.com/TykTechnologies/tyk-identity-broker

go 1.19

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/TykTechnologies/storage v1.0.5
	github.com/crewjam/saml v0.4.12
	github.com/go-ldap/ldap/v3 v3.2.3
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/markbates/goth v1.64.2
	github.com/matryer/is v1.4.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.7
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/crewjam/httperr v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/mrjones/oauth v0.0.0-20180629183705-f4e24b6d100c // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russellhaering/goxmldsig v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.11.2 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220906165146-f3363e06e74c // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jeffail/tunny => github.com/Jeffail/tunny v0.0.0-20171107125207-452a8e97d6a3

replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20211213120648-56cd4003725b

replace gorm.io/gorm => github.com/TykTechnologies/gorm v1.20.7-0.20210409171139-b5c340f85ed0

exclude github.com/TykTechnologies/tyk/certs v0.0.1
