package models

import (
	"time"

	"github.com/sk25469/kv/internal/utils"
)

// KeyValue represents a key-value pair
type KeyValue struct {
	Value      string `json:"value"`
	expiration time.Time
}

func NewKeyValue() *KeyValue {
	return &KeyValue{
		expiration: utils.INFINITY,
	}
}

func (kv *KeyValue) SetExpiration(ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	kv.expiration = expiration
}

func (kv *KeyValue) GetExpiration() time.Time {
	return kv.expiration
}
