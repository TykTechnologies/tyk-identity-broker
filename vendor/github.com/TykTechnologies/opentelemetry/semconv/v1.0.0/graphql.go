package semconv

import (
	"github.com/TykTechnologies/opentelemetry/trace"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// GraphQLPrefix is the base prefix for all GraphQL attributes
	GraphQLPrefix = "graphql."
	// GraphQLOperationPrefix is the base prefix for all the GraphQL operation attributes
	GraphQLOperationPrefix = GraphQLPrefix + "operation."
)

const (
	// GraphQLOperationNameKey represents the name of the operation being executed.
	GraphQLOperationNameKey = attribute.Key(GraphQLOperationPrefix + "name")

	// GraphQLOperationTypeKey The type of the operation being executed.
	GraphQLOperationTypeKey = attribute.Key(GraphQLOperationPrefix + "type")
)

const (
	// GraphQLDocumentKey represents The GraphQL document being executed.
	GraphQLDocumentKey = attribute.Key(GraphQLPrefix + "document")
)

// GraphQLOperationName returns an attribute KeyValue conforming to the
// "operation.name" semantic convention.
func GraphQLOperationName(name string) trace.Attribute {
	return GraphQLOperationNameKey.String(name)
}

// GraphQLOperationType returns an attribute KeyValue conforming to the
// "operation.type" semantic convention.
func GraphQLOperationType(operationType string) trace.Attribute {
	return GraphQLOperationTypeKey.String(operationType)
}

// GraphQLDocument returns an attribute KeyValue conforming to the
// "document" semantic convention.
func GraphQLDocument(document string) trace.Attribute {
	return GraphQLDocumentKey.String(document)
}
