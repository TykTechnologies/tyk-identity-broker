package trace

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

type Attribute = attribute.KeyValue

// NewAttribute creates a new attribute.KeyValue pair based on the provided key and value.
// The function supports multiple types for the value parameter including
// basic types (string, bool, int, int64, float64), their pointer types, slices of basic types,
// and any type implementing the fmt.Stringer interface.
//
// Usage:
//
//	attr := trace.NewAttribute("key1", "value1")
//	fmt.Println(attr) // Output: "key1":"value1"
func NewAttribute(key string, value interface{}) Attribute {
	switch v := value.(type) {
	case string:
		return attribute.Key(key).String(v)
	case *string:
		return attribute.Key(key).String(*v)
	case bool:
		return attribute.Key(key).Bool(v)
	case *bool:
		return attribute.Key(key).Bool(*v)
	case int:
		return attribute.Key(key).Int(v)
	case *int:
		return attribute.Key(key).Int(*v)
	case int64:
		return attribute.Key(key).Int64(v)
	case *int64:
		return attribute.Key(key).Int64(*v)
	case float64:
		return attribute.Key(key).Float64(v)
	case *float64:
		return attribute.Key(key).Float64(*v)
	case []string:
		return attribute.Key(key).StringSlice(v)
	case []bool:
		return attribute.Key(key).BoolSlice(v)
	case []int:
		return attribute.Key(key).IntSlice(v)
	case []int64:
		return attribute.Key(key).Int64Slice(v)
	case []float64:
		return attribute.Key(key).Float64Slice(v)
	case fmt.Stringer:
		return attribute.Key(key).String(v.String())
	default:
		return attribute.Key(key).String(fmt.Sprint(v))
	}
}
