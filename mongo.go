package decimal

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

func Encoder(encodeContext bsoncodec.EncodeContext, writer bsonrw.ValueWriter, value reflect.Value) error {
	return writer.WriteString(value.Interface().(Decimal).String())
}

func Decoder(decodeContext bsoncodec.DecodeContext, reader bsonrw.ValueReader, value reflect.Value) error {
	str, err := reader.ReadString()
	if err != nil {
		return err
	}
	dec, err := NewFromString(str)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(dec))
	return nil
}
