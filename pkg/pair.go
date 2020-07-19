package bt

import (
	"bytes"
	"fmt"
)

// KeyType is the key type in b-tree
type KeyType []byte

// ValueType is the value type in b-tree
type ValueType []byte

// pair represents key-value pair
type pair struct {
	key   KeyType
	value ValueType
}

func (key KeyType) lessThan(other KeyType) bool {
	return bytes.Compare(key, other) == -1
}

func (key KeyType) equalTo(other KeyType) bool {
	return bytes.Compare(key, other) == 0
}

func (key KeyType) greaterThan(other KeyType) bool {
	return bytes.Compare(key, other) == 1
}

// newPair creates a new pair
func newPair(key KeyType, value ValueType) pair {
	return pair{
		key:   key,
		value: value,
	}
}

func (p pair) String() string {
	return fmt.Sprintf("%s=%s", p.key, p.value)
}
