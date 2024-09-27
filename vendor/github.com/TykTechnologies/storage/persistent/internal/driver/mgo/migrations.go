package mgo

import "gopkg.in/mgo.v2"

func buildOpt(opt map[string]interface{}) *mgo.CollectionInfo {
	collectionInfo := &mgo.CollectionInfo{}

	if val, ok := opt["capped"].(bool); ok {
		collectionInfo.Capped = val
	}

	if val, ok := opt["maxBytes"].(int); ok {
		collectionInfo.MaxBytes = val
	}

	if val, ok := opt["maxDocs"].(int); ok {
		collectionInfo.MaxDocs = val
	}

	if val, ok := opt["disableIdIndex"].(bool); ok {
		collectionInfo.DisableIdIndex = val
	}

	if val, ok := opt["forceIdIndex"].(bool); ok {
		collectionInfo.ForceIdIndex = val
	}

	if val, ok := opt["validator"]; ok {
		collectionInfo.Validator = val
	}

	if val, ok := opt["validationLevel"].(string); ok {
		collectionInfo.ValidationLevel = val
	}

	if val, ok := opt["validationAction"].(string); ok {
		collectionInfo.ValidationAction = val
	}

	if val, ok := opt["storageEngine"].(string); ok {
		collectionInfo.StorageEngine = val
	}

	return collectionInfo
}
