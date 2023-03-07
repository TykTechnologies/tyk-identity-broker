module github.com/TykTechnologies/tyk-identity-broker

go 1.19

require (
	github.com/Jeffail/gabs v1.4.0
	github.com/TykTechnologies/storage v0.0.0-20230306183208-bb5bfcd14c9c
	github.com/TykTechnologies/tyk v1.9.2-0.20211217130848-b04d51712be7
	github.com/crewjam/saml v0.4.6
	github.com/go-ldap/ldap/v3 v3.2.3
	github.com/go-redis/redis/v8 v8.3.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/sessions v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/markbates/goth v1.64.2
	github.com/matryer/is v1.4.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.3-0.20191026113918-67a7fdcf741f
	github.com/stretchr/testify v1.8.1
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.7
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/TykTechnologies/gojsonschema v0.0.0-20170222154038-dcb3e4bb7990 // indirect
	github.com/TykTechnologies/murmur3 v0.0.0-20180602122059-1915e687e465 // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/crewjam/httperr v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eclipse/paho.mqtt.golang v1.2.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jensneuse/abstractlogger v0.0.4 // indirect
	github.com/jensneuse/graphql-go-tools v1.20.2 // indirect
	github.com/jensneuse/pipeline v0.0.0-20200117120358-9fb4de085cd6 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lonelycode/go-uuid v0.0.0-20141202165402-ed3ca8a15a93 // indirect
	github.com/lonelycode/osin v0.0.0-20160423095202-da239c9dacb6 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/mrjones/oauth v0.0.0-20180629183705-f4e24b6d100c // indirect
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pmylund/go-cache v2.1.0+incompatible // indirect
	github.com/russellhaering/goxmldsig v1.1.1 // indirect
	github.com/tidwall/gjson v1.11.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/sjson v1.0.4 // indirect
	github.com/uber/jaeger-client-go v2.19.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.11.2 // indirect
	go.opentelemetry.io/otel v0.13.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.18.1 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/gorm v1.21.11 // indirect
)

replace github.com/crewjam/saml => github.com/TykTechnologies/saml v0.4.6-0.20211129150050-bf5ed7a9f748

replace github.com/jeffail/tunny => github.com/Jeffail/tunny v0.0.0-20171107125207-452a8e97d6a3

replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20211213120648-56cd4003725b

replace gorm.io/gorm => github.com/TykTechnologies/gorm v1.20.7-0.20210409171139-b5c340f85ed0

exclude github.com/TykTechnologies/tyk/certs v0.0.1
