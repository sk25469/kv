package models

import (
	"time"

	"github.com/sk25469/kv/utils"
)

// Value represents a key-value pair
type Value struct {
	Value      string `json:"value"`
	expiration time.Time
}

func NewKeyValue(val string) *Value {
	return &Value{
		Value:      val,
		expiration: utils.INFINITY,
	}
}

func (kv *Value) SetExpiration(ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	kv.expiration = expiration
}

func (kv *Value) GetExpiration() time.Time {
	return kv.expiration
}
