package model

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type ObjectID string

func NewObjectID() ObjectID {
	return ObjectID(bson.NewObjectId())
}

func NewObjectIDWithTime(t time.Time) ObjectID {
	return ObjectID(bson.NewObjectIdWithTime(t))
}

// Valid returns true if id is valid. A valid id must contain exactly 12 bytes.
func (id ObjectID) Valid() bool {
	return len(id) == 12
}

func (id ObjectID) Hex() string {
	return hex.EncodeToString([]byte(id))
}

func (id ObjectID) String() string {
	return id.Hex()
}

func (id ObjectID) Timestamp() time.Time {
	bytes := []byte(string(id)[0:4])
	secs := int64(binary.BigEndian.Uint32(bytes))

	return time.Unix(secs, 0)
}

func (id ObjectID) Time() time.Time {
	return id.Timestamp()
}

func (id ObjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Hex())
}

func (id *ObjectID) UnmarshalJSON(buf []byte) error {
	var b bson.ObjectId
	err := b.UnmarshalJSON(buf)

	*id = ObjectID(string(b))

	return err
}

// ObjectIDHex useful to create an object ID from the string
func ObjectIDHex(id string) ObjectID {
	return ObjectID(bson.ObjectIdHex(id))
}

func IsObjectIDHex(s string) bool {
	if len(s) != 24 {
		return false
	}

	_, err := hex.DecodeString(s)

	return err == nil
}

// GetBSON only used by mgo
func (id ObjectID) GetBSON() (interface{}, error) {
	return bson.ObjectId(id), nil
}

// Value is being used by SQL drivers
func (id ObjectID) Value() (driver.Value, error) {
	return bson.ObjectId(id).Hex(), nil
}

func (id *ObjectID) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}

	// reflect magic to update existing string without creating new one
	if len(bytes) > 0 {
		bs := ObjectID(bson.ObjectIdHex(string(bytes)))
		*id = bs
	}

	return nil
}
