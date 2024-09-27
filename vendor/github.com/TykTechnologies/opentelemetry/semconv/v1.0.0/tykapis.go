package semconv

import (
	"github.com/TykTechnologies/opentelemetry/trace"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// TykAPIPrefix is the base prefix for all the TykAPIS attributes
	TykAPIPrefix = "tyk.api."
)

// Attributes that should be present on all the Tyk Gateway API traces
const (
	// equivalent to Tyk Pump APIID (api_id)
	TykAPIIDKey = attribute.Key(TykAPIPrefix + "id")

	// equivalent to Tyk Pump APIName(api_name)
	TykAPINameKey = attribute.Key(TykAPIPrefix + "name")

	// equivalent to Tyk Pump OrgID(org_id)
	TykAPIOrgIDKey = attribute.Key(TykAPIPrefix + "orgid")

	// equivalent to Tyk Pump Tags(tags)
	TykAPITagsKey = attribute.Key(TykAPIPrefix + "tags")

	// equivalent to Tyk Pump API Listen Path (path/raw_path)
	TykAPIListenPathKey = attribute.Key(TykAPIPrefix + "path")
)

// Version related attributes
const (
	// equivalent to Tyk Pump APIVersion(api_version)
	TykAPIVersionKey = attribute.Key(TykAPIPrefix + "version")
)

// Auth related attributes
const (
	// equivalent to Tyk Pump APIKey(api_key)
	TykAPIKeyKey = attribute.Key(TykAPIPrefix + "apikey")

	// equivalent to Tyk Pump Alias(alias)
	TykAPIKeyAliasKey = attribute.Key(TykAPIPrefix + "apikey.alias")

	// equivalent to Tyk Pump OauthID(oauth_id)
	TykOauthIDKey = attribute.Key(TykAPIPrefix + "oauthid")
)

// TykAPIID returns an attribute KeyValue conforming to the
// "tyk.api.id" semantic convention. It represents the id
// of the Tyk API.
func TykAPIID(id string) trace.Attribute {
	return TykAPIIDKey.String(id)
}

// TykAPIName returns an attribute KeyValue conforming to the
// "tyk.api.name" semantic convention. It represents the name
// of the Tyk API.
func TykAPIName(name string) trace.Attribute {
	return TykAPINameKey.String(name)
}

// TykAPIVersion returns an attribute KeyValue conforming to the
// "tyk.api.version" semantic convention. It represents the version
// of the Tyk API.
func TykAPIVersion(version string) trace.Attribute {
	return TykAPIVersionKey.String(version)
}

// TykAPIOrgIDKey returns an attribute KeyValue conforming to the
// "tyk.api.orgid" semantic convention. It represents the org_id
// of the Tyk API.
func TykAPIOrgID(orgid string) trace.Attribute {
	return TykAPIOrgIDKey.String(orgid)
}

// TykAPIListenPath returns an attribute KeyValue conforming to the
// "tyk.api.path" semantic convention. It represents the listen path
// of the Tyk API.
func TykAPIListenPath(path string) trace.Attribute {
	return TykAPIListenPathKey.String(path)
}

// TykAPITags returns an attribute KeyValue conforming to the
// "tyk.api.tags" semantic convention. It represents the session context
// tags. Can contain many tags which refer to many things, such as the gateway,
// API key, organisation, API definition etc. Concatenated by space.
func TykAPITags(tags ...string) trace.Attribute {
	return TykAPITagsKey.StringSlice(tags)
}

// TykAPIKey returns an attribute KeyValue conforming to the
// "tyk.api.apikey" semantic convention. It represents the authentication
// key for the request.
func TykAPIKey(key string) trace.Attribute {
	return TykAPIKeyKey.String(key)
}

// TykAPIKeyAlias returns an attribute KeyValue conforming to the
// "tyk.api.apikey.alias" semantic convention. It represents the api key
// alias. Blank if no alias is set or request is unauthenticated.
func TykAPIKeyAlias(alias string) trace.Attribute {
	return TykAPIKeyAliasKey.String(alias)
}

// TykOauthID returns an attribute KeyValue conforming to the
// "tyk.api.oauthid" semantic convention. It represents the id of the Oauth Client.
// Value is empty string if not using OAuth, or OAuth client not present.
func TykOauthID(oauthID string) trace.Attribute {
	return TykOauthIDKey.String(oauthID)
}
