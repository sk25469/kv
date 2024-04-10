package models

import (
	"net"
	"time"

	"github.com/sk25469/kv/utils"
)

type ClientConfig struct {
	ClientID    string
	IPAddress   string
	ConnectTime time.Time
	ClientState *ClientState
	Connection  *net.Conn
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
