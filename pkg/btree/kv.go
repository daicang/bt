package btree

import (
	"bytes"
	"fmt"
)

type keyType []byte

func (key keyType) lessThan(other keyType) bool {
	return bytes.Compare(key, other) == -1
}

func (key keyType) equalTo(other keyType) bool {
	return bytes.Compare(key, other) == 0
}

func (key keyType) greaterThen(other keyType) bool {
	return bytes.Compare(key, other) == 1
}

// KV is key-value type stored in btree
type KV struct {
	key   keyType
	value []byte
}

// newKV creates a new KV
func newKV(key, value []byte) KV {
	return KV{
		key:   key,
		value: value,
	}
}

func (kv KV) String() string {
	return fmt.Sprintf("%s=%s", kv.key, kv.value)
}
