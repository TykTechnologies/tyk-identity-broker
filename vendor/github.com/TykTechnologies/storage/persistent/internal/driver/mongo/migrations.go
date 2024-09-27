package mongo

import (
	"github.com/TykTechnologies/storage/persistent/model"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func buildOpt(opt model.DBM) *options.CreateCollectionOptions {
	opts := options.CreateCollection()

	if val, ok := opt["capped"].(bool); ok {
		opts.SetCapped(val)
	}

	if val, ok := opt["collation"].(model.DBM); ok {
		opts.SetCollation(buildCollation(val))
	}

	if val, ok := opt["maxDocs"].(int); ok {
		opts.SetMaxDocuments(int64(val))
	}

	if val, ok := opt["maxBytes"].(int); ok {
		opts.SetSizeInBytes(int64(val))
	}

	if val, ok := opt["storageEngine"]; ok {
		opts.SetChangeStreamPreAndPostImages(val)
	}

	if val, ok := opt["validationAction"].(string); ok {
		opts.SetValidationAction(val)
	}

	if val, ok := opt["validationLevel"].(string); ok {
		opts.SetValidationLevel(val)
	}

	if val, ok := opt["validator"]; ok {
		opts.SetValidator(val)
	}

	if val, ok := opt["expireAfterSeconds"].(int); ok {
		opts.SetExpireAfterSeconds(int64(val))
	}

	if val, ok := opt["timeSeries"].(model.DBM); ok {
		opts.SetTimeSeriesOptions(buildTimeSeriesOptions(val))
	}

	if val, ok := opt["encryptedFields"]; ok {
		opts.SetEncryptedFields(val)
	}

	if val, ok := opt["clusteredIndex"]; ok {
		opts.SetClusteredIndex(val)
	}

	return opts
}

func buildCollation(collation model.DBM) *options.Collation {
	opts := options.Collation{}

	if val, ok := collation["locale"].(string); ok {
		opts.Locale = val
	}

	if val, ok := collation["caseLevel"].(bool); ok {
		opts.CaseLevel = val
	}

	if val, ok := collation["caseFirst"].(string); ok {
		opts.CaseFirst = val
	}

	if val, ok := collation["strength"].(int); ok {
		opts.Strength = val
	}

	if val, ok := collation["numericOrdering"].(bool); ok {
		opts.NumericOrdering = val
	}

	if val, ok := collation["alternate"].(string); ok {
		opts.Alternate = val
	}

	if val, ok := collation["maxVariable"].(string); ok {
		opts.MaxVariable = val
	}

	if val, ok := collation["normalization"].(bool); ok {
		opts.Normalization = val
	}

	if val, ok := collation["backwards"].(bool); ok {
		opts.Backwards = val
	}

	return &opts
}

func buildTimeSeriesOptions(timeSeries model.DBM) *options.TimeSeriesOptions {
	opts := options.TimeSeriesOptions{}

	if val, ok := timeSeries["timeField"].(string); ok {
		opts.SetTimeField(val)
	}

	if val, ok := timeSeries["metaField"].(string); ok {
		opts.SetMetaField(val)
	}

	if val, ok := timeSeries["granularity"].(string); ok {
		opts.SetGranularity(val)
	}

	return &opts
}
