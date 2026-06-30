package stream

import (
	"iter"
	"reflect"
)

// BingenStream is the stream interface for all streamable types
type BingenStream interface {
	// Stream returns the iterator which will stream each field of the target type and
	// return the field info as well as the value.
	Stream() iter.Seq2[BingenFieldInfo, *BingenValue]

	// Close will close any dynamic io.Reader used to stream in the fields
	Close()

	// Error returns an error if one occurred during the process of streaming the type's fields.
	// This can be checked after iterating through the Stream().
	Error() error
}

// BingenValue contains the value of a field as well as any index/key associated with that value.
type BingenValue struct {
	Value any
	Index any
}

// IsNil is just a method accessor way to check to see if the value returned was nil
func (bv *BingenValue) IsNil() bool {
	return bv == nil
}

// SingleV creates a single BingenValue instance without a key or index
func SingleV(value any) *BingenValue {
	return &BingenValue{
		Value: value,
	}
}

// PairV creates a pair of key/index and value.
func PairV(index any, value any) *BingenValue {
	return &BingenValue{
		Value: value,
		Index: index,
	}
}

// BingenFieldInfo contains the type of the field being streamed as well as the name of the field.
type BingenFieldInfo struct {
	Type reflect.Type
	Name string
}
