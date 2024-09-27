package types

import (
	"context"

	"github.com/TykTechnologies/storage/persistent/model"
	"github.com/TykTechnologies/storage/persistent/utils"
)

type PersistentStorage interface {
	// Insert a DbObject into the database
	Insert(context.Context, ...model.DBObject) error
	// Delete a DbObject from the database
	Delete(context.Context, model.DBObject, ...model.DBM) error
	// Update a DbObject in the database
	Update(context.Context, model.DBObject, ...model.DBM) error
	// Count counts all rows for a DBTable if no filter model.DBM given.
	// If a filter model.DBM is specified it will count the rows given the built query for that filter.
	// If multiple filters model.DBM are specified, it will return an error.
	// In case of an error, the count result is going to be 0.
	Count(ctx context.Context, row model.DBObject, filter ...model.DBM) (count int, error error)
	// Query one or multiple DBObjects from the database
	Query(context.Context, model.DBObject, interface{}, model.DBM) error
	// BulkUpdate updates multiple rows
	BulkUpdate(context.Context, []model.DBObject, ...model.DBM) error
	// UpdateAll executes the update query model.DBM over
	// the elements filtered by query model.DBM in the row model.DBObject collection
	UpdateAll(ctx context.Context, row model.DBObject, query, update model.DBM) error
	// Drop drops the collection given the TableName() of the model.DBObject
	Drop(context.Context, model.DBObject) error
	// CreateIndex creates an model.Index in row model.DBObject TableName()
	CreateIndex(ctx context.Context, row model.DBObject, index model.Index) error
	// GetIndexes returns all the model.Index associated to row model.DBObject
	GetIndexes(ctx context.Context, row model.DBObject) ([]model.Index, error)
	// Ping checks if the database is reachable
	Ping(context.Context) error
	// HasTable checks if the table/collection exists
	HasTable(context.Context, string) (bool, error)
	// DropDatabase removes the database
	DropDatabase(ctx context.Context) error
	// Migrate creates the table/collection if it doesn't exist
	Migrate(context.Context, []model.DBObject, ...model.DBM) error
	// DBTableStats retrieves statistics for a specified table in the database.
	// The function takes a context.Context and an model.DBObject as input parameters,
	// where the DBObject represents the table to get stats for.
	// The result is decoded into a model.DBM object, along with any error that occurred during the command execution.
	// Example: stats["capped"] -> true
	DBTableStats(ctx context.Context, row model.DBObject) (model.DBM, error)
	// Aggregate performs an aggregation query on the row model.DBObject collection
	// query is the aggregation pipeline to be executed
	// it returns the aggregation result and an error if any
	Aggregate(ctx context.Context, row model.DBObject, query []model.DBM) ([]model.DBM, error)
	// CleanIndexes removes all the indexes from the row model.DBObject collection
	CleanIndexes(ctx context.Context, row model.DBObject) error
	// Upsert performs an upsert operation on the row model.DBObject collection
	// query is the filter to be used to find the document to update
	// update is the update to be applied to the document
	// row is modified with the result of the operation
	Upsert(ctx context.Context, row model.DBObject, query, update model.DBM) error
	// GetDatabaseInfo returns information of the database to which the driver is connecting to
	GetDatabaseInfo(ctx context.Context) (utils.Info, error)
	// GetTables return the list of collections for a given database
	GetTables(ctx context.Context) ([]string, error)
	// DropTable drops a table/collection from the database. Returns the number of affected rows and error
	DropTable(ctx context.Context, name string) (int, error)
}
