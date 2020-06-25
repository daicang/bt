package bt

import (
	"bytes"
	"fmt"
)

type keyType []byte
type valueType []byte

// KV is key-value type stored in btree
type KV struct {
	key   keyType
	value valueType
}

func (key keyType) lessThan(other keyType) bool {
	return bytes.Compare(key, other) == -1
}

func (key keyType) equalTo(other keyType) bool {
	return bytes.Compare(key, other) == 0
}

func (key keyType) greaterThan(other keyType) bool {
	return bytes.Compare(key, other) == 1
}

// newKV creates a new KV
func newKV(key keyType, value []byte) KV {
	return KV{
		key:   key,
		value: value,
	}
}

func (kv KV) String() string {
	return fmt.Sprintf("%s=%s", kv.key, kv.value)
}
