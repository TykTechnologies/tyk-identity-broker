package mongo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/TykTechnologies/storage/persistent/internal/helper"
	"github.com/TykTechnologies/storage/persistent/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/TykTechnologies/storage/persistent/internal/types"
)

type lifeCycle struct {
	client *mongo.Client

	connectionString string
	database         string
}

var _ types.StorageLifecycle = &lifeCycle{}

const (
	MongoPrefix    = "mongodb://"
	MongoSRVPrefix = "mongodb+srv://"
)

// Connect connects to the mongo database given the ClientOpts.
func (lc *lifeCycle) Connect(opts *types.ClientOpts) error {
	var err error
	var client *mongo.Client

	url, cs, err := parseURL(opts.ConnectionString)
	if err != nil {
		return err
	}

	opts.ConnectionString = url

	connOpts, err := mongoOptsBuilder(opts)
	if err != nil {
		return errors.New(err.Error())
	}

	// SetRegistry allow us to marshall/unmarshall old mgo ID's structures and mgo default values.
	connOpts.SetRegistry(createCustomRegistry().Build())

	if client, err = mongo.Connect(context.Background(), connOpts); err != nil {
		return err
	}

	lc.connectionString = opts.ConnectionString
	lc.database = cs.db
	lc.client = client

	return lc.client.Ping(context.Background(), nil)
}

type urlInfo struct {
	addrs   []string
	user    string
	pass    string
	db      string
	options []urlOptions
}

// urlOptions is a key/value pair representing a single option in a URL.
// we need to use this struct instead of a map to avoid flaky tests due to the order of the options
type urlOptions struct {
	key string
	val string
}

func isOptSep(c rune) bool {
	return c == ';' || c == '&'
}

func parseURL(s string) (string, *urlInfo, error) {
	var info *urlInfo
	prefix := ""

	if strings.HasPrefix(s, MongoPrefix) {
		prefix = MongoPrefix
	} else if strings.HasPrefix(s, MongoSRVPrefix) {
		prefix = MongoSRVPrefix
	}

	switch prefix {
	case MongoPrefix:
		s = strings.TrimPrefix(s, MongoPrefix)
	case MongoSRVPrefix:
		s = strings.TrimPrefix(s, MongoSRVPrefix)
	default:
		return "", info, errors.New("invalid connection string, no prefix found")
	}

	info, err := extractURL(s)
	if err != nil {
		return "", info, err
	}

	var connString string
	connString += prefix

	if info.user != "" {
		info.user = url.QueryEscape(info.user)
		connString += info.user

		if info.pass != "" {
			info.pass = url.QueryEscape(info.pass)
			connString += ":" + info.pass
		}

		connString += "@"
	}

	connString += strings.Join(info.addrs, ",")

	connString += "/" + info.db

	if len(info.options) > 0 {
		connString += "?"
		for _, v := range info.options {
			connString += v.key + "=" + v.val + "&"
		}

		connString = connString[:len(connString)-1]
	}

	return connString, info, nil
}

func extractURL(s string) (*urlInfo, error) {
	info := &urlInfo{options: make([]urlOptions, 0)}
	var err error

	if s, err = extractOptions(s, info); err != nil {
		return nil, err
	}

	if s, err = extractCredentials(s, info); err != nil {
		return nil, err
	}

	if s, err = extractDatabase(s, info); err != nil {
		return nil, err
	}

	info.addrs = strings.Split(s, ",")

	return info, nil
}

func extractOptions(s string, info *urlInfo) (string, error) {
	if c := strings.Index(s, "?"); c != -1 {
		for _, pair := range strings.FieldsFunc(s[c+1:], isOptSep) {
			l := strings.SplitN(pair, "=", 2)
			if len(l) != 2 || l[0] == "" || l[1] == "" {
				return s, errors.New("connection option must be key=value: " + pair)
			}

			info.options = append(info.options, urlOptions{key: l[0], val: l[1]})
		}

		s = s[:c]
	}

	return s, nil
}

func extractCredentials(s string, info *urlInfo) (string, error) {
	if c := strings.Index(s, "@"); c != -1 {
		pair := strings.SplitN(s[:c], ":", 2)
		if len(pair) > 2 || pair[0] == "" {
			return s, errors.New("credentials must be provided as user:pass@host")
		}

		var err error

		info.user, err = url.QueryUnescape(pair[0])
		if err != nil {
			return s, fmt.Errorf("cannot unescape username in URL: %q", pair[0])
		}

		if len(pair) > 1 {
			info.pass, err = url.QueryUnescape(pair[1])
			if err != nil {
				return s, fmt.Errorf("cannot unescape password in URL")
			}
		}

		s = s[c+1:]
	}

	return s, nil
}

func extractDatabase(s string, info *urlInfo) (string, error) {
	if c := strings.Index(s, "/"); c != -1 {
		info.db = s[c+1:]
		s = s[:c]
	}

	return s, nil
}

// Close finish the session.
func (lc *lifeCycle) Close() error {
	if lc.client != nil {
		return lc.client.Disconnect(context.Background())
	}

	return errors.New("closing a no connected database")
}

// DBType returns the type of the registered storage driver.
func (lc *lifeCycle) DBType() utils.DBType {
	if helper.IsCosmosDB(lc.connectionString) {
		return utils.CosmosDB
	}

	var result struct {
		Code int `bson:"code"`
	}

	cmd := bson.D{{Key: "features", Value: 1}}
	singleResult := lc.client.Database("admin").RunCommand(context.Background(), cmd)

	if err := singleResult.Decode(&result); (singleResult.Err() != nil || err != nil) && result.Code == 303 {
		return utils.AWSDocumentDB
	}

	return utils.StandardMongo
}

// mongoOptsBuilder build Mongo options.ClientOptions from our own types.ClientOpts. Also sets default values.
// mongo URI parameters specified in the types.ClientOpts ConnectionString have precedence over the ones configured in
// other input.
func mongoOptsBuilder(opts *types.ClientOpts) (*options.ClientOptions, error) {
	connOpts := options.Client()

	if opts.UseSSL {
		tlsConfig, err := opts.GetTLSConfig()
		if err != nil {
			return nil, err
		}

		connOpts.SetTLSConfig(tlsConfig)
	}

	connOpts.SetTimeout(types.DEFAULT_CONN_TIMEOUT)

	if opts.ConnectionTimeout != 0 {
		connOpts.SetTimeout(time.Duration(opts.ConnectionTimeout) * time.Second)
	}

	connOpts.SetReadPreference(getReadPrefFromConsistency(opts.SessionConsistency))

	// we apply URI here so if we specify a different configuration in the URI it can be overridden
	connOpts.ApplyURI(opts.ConnectionString)

	connOpts.SetDirect(opts.DirectConnection)

	err := connOpts.Validate()
	if err != nil {
		return nil, err
	}

	return connOpts, nil
}

// getReadPrefFromConsistency returns the equivalent of the readPreference for session consistency
func getReadPrefFromConsistency(consistency string) *readpref.ReadPref {
	var mode *readpref.ReadPref

	switch consistency {
	case "eventual":
		mode = readpref.Nearest()
	case "monotonic":
		mode = readpref.PrimaryPreferred()
	default:
		mode = readpref.Primary()
	}

	return mode
}
