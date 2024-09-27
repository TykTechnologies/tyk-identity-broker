package mongo

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/TykTechnologies/storage/persistent/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildLimitQuery(fields ...string) bson.D {
	order := bson.D{}

	for _, field := range fields {
		if field == "" {
			continue
		}

		n := 1
		var kind string

		if field != "" {
			if field[0] == '$' {
				if c := strings.Index(field, ":"); c > 1 && c < len(field)-1 {
					kind = field[1:c]
					field = field[c+1:]
				}
			}

			switch field[0] {
			case '+':
				field = field[1:]
			case '-':
				n = -1
				field = field[1:]
			}
		}

		if kind == "textScore" {
			order = append(order, primitive.E{Key: field, Value: bson.M{"$meta": kind}})
		} else {
			order = append(order, primitive.E{Key: field, Value: n})
		}
	}

	return order
}

func handleQueryValue(key string, value interface{}, search bson.M) {
	switch {
	case isNestedQuery(value):
		handleNestedQuery(search, key, value)
	case reflect.ValueOf(value).Kind() == reflect.Slice && key != "$or":
		strSlice, isStrSlice := value.([]string)

		if isStrSlice && key == "_id" {
			ObjectIDs := []model.ObjectID{}
			for _, str := range strSlice {
				if primitive.IsValidObjectID(str) {
					ObjectIDs = append(ObjectIDs, model.ObjectIDHex(str))
				}
			}

			search[key] = bson.M{"$in": ObjectIDs}

			return
		}

		search[key] = primitive.M{"$in": value}
	default:
		search[key] = value
	}
}

// isNestedQuery returns true if the value is model.DBM
func isNestedQuery(value interface{}) bool {
	_, ok := value.(model.DBM)
	return ok
}

// handleNestedQuery replace children queries by it nested values.
// For example, transforms a model.DBM{"testName": model.DBM{"$ne": "123"}} to {"testName":{"$ne":"123"}}
func handleNestedQuery(search bson.M, key string, value interface{}) {
	nestedQuery, ok := value.(model.DBM)
	if !ok {
		return
	}

	for nestedKey, nestedValue := range nestedQuery {
		switch nestedKey {
		case "$i":
			if stringValue, ok := nestedValue.(string); ok {
				quoted := regexp.QuoteMeta(stringValue)
				search[key] = &primitive.Regex{Pattern: fmt.Sprintf("^%s$", quoted), Options: "i"}
			}
		case "$text":
			if stringValue, ok := nestedValue.(string); ok {
				search[key] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(stringValue), Options: "i"}}
			}
		default:
			if v, ok := search[key]; !ok {
				search[key] = bson.M{nestedKey: nestedValue}
			} else {
				if nestedQ, ok := v.(bson.M); ok {
					nestedQ[nestedKey] = nestedValue
					search[key] = nestedQ
				}
			}
		}
	}
}

// buildQuery transforms model.DBM into bson.M (primitive.M) it does some special treatment to nestedQueries
// using handleNestedQuery func.
func buildQuery(query model.DBM) bson.M {
	search := bson.M{}

	for key, value := range query {
		switch key {
		case "_sort", "_collection", "_limit", "_offset", "_date_sharding":
			continue
		case "_id":
			if id, ok := value.(model.ObjectID); ok {
				search[key] = id
				continue
			}

			handleQueryValue(key, value, search)
		default:
			handleQueryValue(key, value, search)
		}
	}

	return search
}
