package models

import "github.com/sk25469/kv/internal/utils"

// state can be 1 of the following:
// - transactional
// - active

type ClientState struct {
	State string
}

// NewClientState creates a new instance of ClientState
func NewClientState() *ClientState {
	return &ClientState{
		State: utils.ACTIVE,
	}
}
