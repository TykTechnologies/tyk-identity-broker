package mgo

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/TykTechnologies/storage/persistent/model"
	"gopkg.in/mgo.v2/bson"
)

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

func handleQueryValue(key string, value interface{}, search bson.M) {
	switch {
	case isNestedQuery(value):
		handleNestedQuery(search, key, value)
	case reflect.ValueOf(value).Kind() == reflect.Slice && key != "$or":
		strSlice, isStr := value.([]string)

		if isStr && key == "_id" {
			ObjectIDs := []model.ObjectID{}
			for _, str := range strSlice {
				if bson.IsObjectIdHex(str) {
					ObjectIDs = append(ObjectIDs, model.ObjectIDHex(str))
				}
			}

			search[key] = bson.M{"$in": ObjectIDs}

			return
		}

		search[key] = bson.M{"$in": value}
	default:
		search[key] = value
	}
}

func isNestedQuery(value interface{}) bool {
	_, ok := value.(model.DBM)
	return ok
}

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
				search[key] = &bson.RegEx{Pattern: fmt.Sprintf("^%s$", quoted), Options: "i"}
			}
		case "$text":
			if stringValue, ok := nestedValue.(string); ok {
				search[key] = bson.M{"$regex": bson.RegEx{Pattern: regexp.QuoteMeta(stringValue), Options: "i"}}
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

func getColName(query model.DBM, row model.DBObject) (string, error) {
	colName, ok := query["_collection"].(string)
	if !ok {
		if row == nil {
			return "", errors.New("unable to find collection name")
		}

		colName = row.TableName()
	}

	return colName, nil
}
