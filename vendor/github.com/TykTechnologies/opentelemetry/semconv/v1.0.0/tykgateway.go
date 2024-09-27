package semconv

import (
	"github.com/TykTechnologies/opentelemetry/trace"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// TykAPIPrefix is the base prefix for all the TykAPIS attributes
	TykGWPrefix = "tyk.gw."
)

// Attributes that should be present on all the Tyk Gateway API traces
const (
	// represents the gateway ID
	TykGWIDKey = attribute.Key(TykGWPrefix + "id")

	// represents if the gateway is in a dataplane (edge gateway)
	TykGWDataplaneKey = attribute.Key(TykGWPrefix + "dataplane")

	// represents the group id of the dataplane gateway
	TykDataplaneGWGroupIDKey = attribute.Key(TykGWPrefix + "group.id")

	// represents the group id of the dataplane gateway
	TykGWSegmentTagsKey = attribute.Key(TykGWPrefix + "tags")
)

// TykGWIDKey returns an attribute KeyValue conforming to the
// "tyk.gw.id" semantic convention.  It represents the id
// of the Tyk Gateway.
func TykGWID(id string) trace.Attribute {
	return TykGWIDKey.String(id)
}

// TykGWDataplane returns an attribute KeyValue conforming to the
// "tyk.gw.dataplane" semantic convention.  It represents if the Tyk Gateway
// is dataplane (slave_options.use_rpc=true).
func TykGWDataplane(isDataplane bool) trace.Attribute {
	return TykGWDataplaneKey.Bool(isDataplane)
}

// TykDataplaneGWGroupID returns an attribute KeyValue conforming to the
// "tyk.gw.group.id" semantic convention.  It represents the db_app_conf_options.tags
// of the Tyk Gateway. It only populated if the gateway is dataplane.
func TykDataplaneGWGroupID(groupID string) trace.Attribute {
	return TykDataplaneGWGroupIDKey.String(groupID)
}

// TykGWSegmentTags returns an attribute KeyValue conforming to the
// "tyk.gw.tags" semantic convention.  It represents the slave_options.group_id
// of the Tyk Gateway. It only populated if the gateway is segmented.
func TykGWSegmentTags(tags ...string) trace.Attribute {
	return TykGWSegmentTagsKey.StringSlice(tags)
}
