package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/TykTechnologies/storage/persistent/internal/helper"
	"github.com/TykTechnologies/storage/persistent/internal/types"
	"github.com/TykTechnologies/storage/persistent/model"
	"github.com/TykTechnologies/storage/persistent/utils"
)

var _ types.PersistentStorage = &mongoDriver{}

type mongoDriver struct {
	*lifeCycle
	options *types.ClientOpts
}

// NewMongoDriver returns an instance of the driver official mongo connected to the database.
func NewMongoDriver(opts *types.ClientOpts) (*mongoDriver, error) {
	if opts.ConnectionString == "" {
		return nil, errors.New("can't connect without connection string")
	}

	newDriver := &mongoDriver{}
	newDriver.options = opts

	// create the db life cycle manager
	lc := &lifeCycle{}

	if err := lc.Connect(opts); err != nil {
		return nil, err
	}

	newDriver.lifeCycle = lc

	return newDriver, nil
}

func (d *mongoDriver) Insert(ctx context.Context, rows ...model.DBObject) error {
	if len(rows) == 0 {
		return errors.New(types.ErrorEmptyRow)
	}

	var bulkQuery []mongo.WriteModel

	for _, row := range rows {
		if row.GetObjectID() == "" {
			row.SetObjectID(model.NewObjectID())
		}

		model := mongo.NewInsertOneModel().SetDocument(row)
		bulkQuery = append(bulkQuery, model)
	}

	collection := d.client.Database(d.database).Collection(rows[0].TableName())
	_, err := collection.BulkWrite(ctx, bulkQuery)

	return d.handleStoreError(err)
}

func (d *mongoDriver) Delete(ctx context.Context, row model.DBObject, query ...model.DBM) error {
	if len(query) > 1 {
		return errors.New(types.ErrorMultipleQueryForSingleRow)
	}

	if len(query) == 0 {
		query = append(query, model.DBM{"_id": row.GetObjectID()})
	}

	collection := d.client.Database(d.database).Collection(row.TableName())

	result, err := collection.DeleteMany(ctx, buildQuery(query[0]))

	if err == nil && result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return d.handleStoreError(err)
}

func (d *mongoDriver) Count(ctx context.Context, row model.DBObject, filters ...model.DBM) (int, error) {
	if len(filters) > 1 {
		return 0, errors.New(types.ErrorMultipleDBM)
	}

	filter := bson.M{}
	if len(filters) == 1 {
		filter = buildQuery(filters[0])
	}

	collection := d.client.Database(d.database).Collection(row.TableName())

	count, err := collection.CountDocuments(ctx, filter)

	return int(count), d.handleStoreError(err)
}

func (d *mongoDriver) Query(ctx context.Context, row model.DBObject, result interface{}, query model.DBM) error {
	collection := d.client.Database(d.database).Collection(row.TableName())

	search := buildQuery(query)

	findOpts := options.Find()
	findOneOpts := options.FindOne()

	sort, sortFound := query["_sort"].(string)
	if sortFound && sort != "" {
		sortQuery := buildLimitQuery(sort)
		findOpts.SetSort(sortQuery)
		findOneOpts.SetSort(sortQuery)
	}

	if limit, ok := query["_limit"].(int); ok && limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	if offset, ok := query["_offset"].(int); ok && offset > 0 {
		findOpts.SetSkip(int64(offset))
		findOneOpts.SetSkip(int64(offset))
	}

	var err error

	if helper.IsSlice(result) {
		var cursor *mongo.Cursor

		cursor, err = collection.Find(ctx, search, findOpts)
		if err == nil {
			err = cursor.All(ctx, result)
			defer cursor.Close(ctx)
		}
	} else {
		err = collection.FindOne(ctx, search, findOneOpts).Decode(result)
	}

	return d.handleStoreError(err)
}

func (d *mongoDriver) Drop(ctx context.Context, row model.DBObject) error {
	collection := d.client.Database(d.database).Collection(row.TableName())

	return d.handleStoreError(collection.Drop(ctx))
}

func (d *mongoDriver) Update(ctx context.Context, row model.DBObject, query ...model.DBM) error {
	if len(query) > 1 {
		return errors.New(types.ErrorMultipleQueryForSingleRow)
	}

	if len(query) == 0 {
		query = append(query, model.DBM{"_id": row.GetObjectID()})
	}

	collection := d.client.Database(d.database).Collection(row.TableName())

	result, err := collection.UpdateMany(ctx, buildQuery(query[0]), bson.D{{Key: "$set", Value: row}})
	if err == nil && result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return d.handleStoreError(err)
}

func (d *mongoDriver) BulkUpdate(ctx context.Context, rows []model.DBObject, query ...model.DBM) error {
	if len(query) > 0 && len(query) != len(rows) {
		return errors.New(types.ErrorRowQueryDiffLenght)
	}

	if len(rows) == 0 {
		return errors.New(types.ErrorEmptyRow)
	}

	var bulkQuery []mongo.WriteModel

	for i := range rows {
		update := mongo.NewUpdateOneModel().SetUpdate(bson.D{{Key: "$set", Value: rows[i]}})

		if len(query) == 0 {
			update.SetFilter(model.DBM{"_id": rows[i].GetObjectID()})
		} else {
			update.SetFilter(buildQuery(query[i]))
		}

		bulkQuery = append(bulkQuery, update)
	}

	collection := d.client.Database(d.database).Collection(rows[0].TableName())
	result, err := collection.BulkWrite(ctx, bulkQuery)
	if err == nil && result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return d.handleStoreError(err)
}

func (d *mongoDriver) UpdateAll(ctx context.Context, row model.DBObject, query, update model.DBM) error {
	collection := d.client.Database(d.database).Collection(row.TableName())

	result, err := collection.UpdateMany(ctx, buildQuery(query), buildQuery(update))
	if err == nil && result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return d.handleStoreError(err)
}

func (d *mongoDriver) HasTable(ctx context.Context, collection string) (bool, error) {
	if d.client == nil {
		return false, errors.New(types.ErrorSessionClosed)
	}

	collections, err := d.client.Database(d.database).ListCollectionNames(ctx, bson.M{"name": collection})

	return len(collections) > 0, err
}

func (d *mongoDriver) Ping(ctx context.Context) error {
	return d.handleStoreError(d.client.Ping(ctx, nil))
}

func (d *mongoDriver) handleStoreError(err error) error {
	if err == nil {
		return nil
	}

	// Check for a mongo.ServerError or any of its underlying wrapped errors
	var serverErr mongo.ServerError
	// Check if the error is a network error
	if mongo.IsNetworkError(err) || errors.As(err, &serverErr) {
		// Reconnect to the MongoDB instance
		if connErr := d.Connect(d.options); connErr != nil {
			return errors.New(types.ErrorReconnecting + ": " + connErr.Error() + " after error: " + err.Error())
		}
	}

	return err
}

func (d *mongoDriver) CreateIndex(ctx context.Context, row model.DBObject, index model.Index) error {
	if len(index.Keys) == 0 {
		return errors.New(types.ErrorIndexEmpty)
	} else if len(index.Keys) > 1 && index.IsTTLIndex {
		return errors.New(types.ErrorIndexComposedTTL)
	}

	keys := bson.D{}

	for _, key := range index.Keys {
		builtQuery := buildQuery(key)
		for name, val := range builtQuery {
			keys = append(keys, bson.E{Key: name, Value: val})
		}
	}

	opts := options.Index()

	//nolint:staticcheck
	opts.SetBackground(index.Background)

	if name := index.Name; name != "" {
		opts.SetName(name)
	}

	if index.IsTTLIndex {
		opts.SetExpireAfterSeconds(int32(index.TTL))
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}

	collection := d.client.Database(d.database).Collection(row.TableName())

	_, err := collection.Indexes().CreateOne(ctx, indexModel)

	return d.handleStoreError(err)
}

func (d *mongoDriver) GetIndexes(ctx context.Context, row model.DBObject) ([]model.Index, error) {
	hasTable, err := d.HasTable(ctx, row.TableName())
	if err != nil {
		return nil, d.handleStoreError(err)
	}

	if !hasTable {
		return nil, errors.New(types.ErrorCollectionNotFound)
	}

	collection := d.client.Database(d.database).Collection(row.TableName())

	var indexes []model.Index

	indexesSpec, err := collection.Indexes().ListSpecifications(ctx)
	if err != nil {
		return indexes, d.handleStoreError(err)
	}

	// parse from mongo IndexSpec to our model.Index again
	for _, thisIndex := range indexesSpec {
		bsonKeys := bson.D{}

		if errUnmarshal := bson.Unmarshal(thisIndex.KeysDocument, &bsonKeys); err != nil {
			return indexes, errUnmarshal
		}

		var newKeys []model.DBM

		for _, v := range bsonKeys {
			newKey := model.DBM{}
			newKey[v.Key] = v.Value

			newKeys = append(newKeys, newKey)
		}

		newIndex := model.Index{
			Name: thisIndex.Name,
			Keys: newKeys,
		}

		if TTL := thisIndex.ExpireAfterSeconds; TTL != nil {
			newIndex.TTL = int(*TTL)
			newIndex.IsTTLIndex = true
		}

		indexes = append(indexes, newIndex)
	}

	return indexes, nil
}

func (d *mongoDriver) Migrate(ctx context.Context, rows []model.DBObject, opts ...model.DBM) error {
	if len(opts) > 0 && len(opts) != len(rows) {
		return errors.New(types.ErrorRowOptDiffLenght)
	}

	for i, row := range rows {
		has, err := d.HasTable(ctx, row.TableName())
		if err != nil {
			return errors.New("error looking for table: " + err.Error())
		}

		if !has {
			var err error

			if len(opts) > 0 {
				opt := buildOpt(opts[i])
				err = d.client.Database(d.database).CreateCollection(ctx, row.TableName(), opt)
			} else {
				err = d.client.Database(d.database).CreateCollection(ctx, row.TableName())
			}

			if err != nil {
				return errors.New("error creating table: " + err.Error())
			}
		}
	}

	return nil
}

func (d *mongoDriver) DropDatabase(ctx context.Context) error {
	return d.client.Database(d.database).Drop(ctx)
}

func (d *mongoDriver) DBTableStats(ctx context.Context, row model.DBObject) (model.DBM, error) {
	var stats model.DBM
	err := d.client.Database(d.database).RunCommand(ctx, bson.D{
		{Key: "collStats", Value: row.TableName()},
	}).Decode(&stats)

	return stats, d.handleStoreError(err)
}

func (d *mongoDriver) Aggregate(ctx context.Context, row model.DBObject, query []model.DBM) ([]model.DBM, error) {
	col := d.client.Database(d.database).Collection(row.TableName())

	cursor, err := col.Aggregate(ctx, query)
	if err != nil {
		return nil, d.handleStoreError(err)
	}

	defer cursor.Close(ctx)

	resultSlice := make([]model.DBM, 0)

	for cursor.Next(ctx) {
		var result model.DBM

		err := cursor.Decode(&result)
		if err != nil {
			return nil, d.handleStoreError(err)
		}

		// Parsing _id from primitive.ObjectID to model.ObjectID
		if ObjectID, ok := result["_id"].(primitive.ObjectID); ok {
			result["_id"] = model.ObjectIDHex(ObjectID.Hex())
		}

		resultSlice = append(resultSlice, result)
	}

	if err := cursor.Err(); err != nil {
		return nil, d.handleStoreError(err)
	}

	return resultSlice, nil
}

func (d *mongoDriver) CleanIndexes(ctx context.Context, row model.DBObject) error {
	collection := d.client.Database(d.database).Collection(row.TableName())

	_, err := collection.Indexes().DropAll(ctx)

	return d.handleStoreError(err)
}

func (d *mongoDriver) Upsert(ctx context.Context, row model.DBObject, query, update model.DBM) error {
	coll := d.client.Database(d.database).Collection(row.TableName())

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := coll.FindOneAndUpdate(ctx, query, update, opts).Decode(row)

	return d.handleStoreError(err)
}

func (d *mongoDriver) GetDatabaseInfo(ctx context.Context) (utils.Info, error) {
	var result utils.Info

	database := d.client.Database("admin")
	err := database.RunCommand(context.Background(), bson.D{primitive.E{Key: "buildInfo", Value: 1}}).Decode(&result)
	result.Type = d.lifeCycle.DBType()

	return result, d.handleStoreError(err)
}

func (d *mongoDriver) GetTables(ctx context.Context) ([]string, error) {
	return d.client.Database(d.database).ListCollectionNames(ctx, bson.D{})
}

func (d *mongoDriver) DropTable(ctx context.Context, collectionName string) (int, error) {
	deleteResult, err := d.client.Database(d.database).Collection(collectionName).DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return int(deleteResult.DeletedCount), d.client.Database(d.database).Collection(collectionName).Drop(ctx)
}
