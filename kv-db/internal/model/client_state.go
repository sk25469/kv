package models

import (
	"time"

	"github.com/sk25469/kv/internal/utils"
)

type ClientConfig struct {
	ClientID    string
	IPAddress   string
	ConnectTime time.Time
	ClientState *ClientState
}

// state can be 1 of the following:
// - transactional
// - active

type ClientState struct {
	State           int
	IsAuthenticated bool
}

// NewClientState creates a new instance of ClientState
func NewClientState() *ClientState {
	return &ClientState{
		State:           utils.ACTIVE,
		IsAuthenticated: false,
	}
}
