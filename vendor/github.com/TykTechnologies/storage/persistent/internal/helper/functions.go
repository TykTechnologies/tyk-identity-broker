package helper

import (
	"log"
	"reflect"
	"strings"
)

func IsSlice(o interface{}) bool {
	return reflect.TypeOf(o).Elem().Kind() == reflect.Slice
}

func ErrPrint(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

func IsCosmosDB(connectionString string) bool {
	return strings.Contains(connectionString, ".cosmos.") ||
		strings.HasPrefix(connectionString, "https://") && strings.Contains(connectionString, ".documents.azure.com") ||
		strings.HasPrefix(connectionString, "tcp://") && strings.Contains(connectionString, ".documents.azure.com") ||
		strings.HasPrefix(connectionString, "mongodb://") && strings.Contains(connectionString, ".documents.azure.com") ||
		strings.Contains(connectionString, "AccountEndpoint=") ||
		strings.Contains(connectionString, "AccountKey=")
}
