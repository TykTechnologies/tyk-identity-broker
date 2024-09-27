package mongo

import (
	"reflect"
	"time"

	"github.com/TykTechnologies/storage/persistent/model"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonoptions"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/mgocompat"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// tOID is the type of model.ObjectID
var tOID = reflect.TypeOf(model.NewObjectID())

// toTime is the type of golang time.Time
var toTime = reflect.TypeOf(time.Time{})

// ObjectIDDecodeValue encode Hex value of model.ObjectID into primitive.ObjectID
func ObjectIDEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tOID {
		return bsoncodec.ValueEncoderError{Name: "ObjectIDEncodeValue", Types: []reflect.Type{tOID}, Received: val}
	}

	s := val.Interface().(model.ObjectID).Hex()

	newOID, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return err
	}

	return vw.WriteObjectID(newOID)
}

// ObjectIDDecodeValue decode Hex value of primitive.ObjectID into model.ObjectID
func ObjectIDDecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	ObjectID, err := vr.ReadObjectID()
	if err != nil {
		return err
	}

	newOID := model.ObjectIDHex(ObjectID.Hex())

	if val.CanSet() {
		val.Set(reflect.ValueOf(newOID))
	}

	return nil
}

// createCustomRegistry creates a *bsoncodec.RegistryBuilder for our lifeCycle mongo's client using  ObjectIDDecodeValue
// and ObjectIDEncodeValue as Type Encoder/Decoders for model.ObjectID and time.Time
func createCustomRegistry() *bsoncodec.RegistryBuilder {
	// using mgocompat registry as base type registry
	rb := mgocompat.NewRegistryBuilder()

	// set the model.ObjectID encoders/decoders
	rb.RegisterTypeEncoder(tOID, bsoncodec.ValueEncoderFunc(ObjectIDEncodeValue))
	rb.RegisterTypeDecoder(tOID, bsoncodec.ValueDecoderFunc(ObjectIDDecodeValue))

	// we set the default behavior to use local time zone - the same as mgo does internally.
	UseLocalTimeZone := true
	opts := &bsonoptions.TimeCodecOptions{UseLocalTimeZone: &UseLocalTimeZone}
	// set the time.Time encoders/decoders
	rb.RegisterTypeDecoder(toTime, bsoncodec.NewTimeCodec(opts))
	rb.RegisterTypeEncoder(toTime, bsoncodec.NewTimeCodec(opts))

	return rb
}
