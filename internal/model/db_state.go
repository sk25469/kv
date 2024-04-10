package models

import (
	"log"
	"net"
	"sync"
)

type DbState struct {
	State       []*Config
	Connections map[string]*net.Conn
	mu          sync.Mutex
}

func (db *DbState) PrintConnections() {
	for address, conn := range db.Connections {
		log.Printf("Address: %v  --------- Conn: %v\n", address, conn)

	}
}

func (db *DbState) AddConnection(address string, conn *net.Conn) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Connections[address] = conn
}

func (db *DbState) RemoveConnection(address string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.Connections, address)
}

func NewDbState() *DbState {
	return &DbState{
		State:       make([]*Config, 0),
		Connections: make(map[string]*net.Conn),
	}
}

func (db *DbState) RemoveFailedDb(config *Config) {
	for i, state := range db.State {
		if state.Port == config.Port {
			db.State = append(db.State[:i], db.State[i+1:]...)
			break
		}
	}
}

func (db *DbState) PrintDbState() {
	for _, state := range db.State {
		log.Printf("Port: %v, IsMaster: %v\n", state.Port, state.IsMaster)
	}
}

func (db *DbState) InsertDb(config *Config) {
	db.State = append(db.State, config)
}

func (db *DbState) GetMaster() *Config {
	for _, state := range db.State {
		if state.IsMaster {
			return state
		}
	}
	return &Config{}
}

func (db *DbState) GetSlaves() []*Config {
	var slaves []*Config
	for _, state := range db.State {
		if !state.IsMaster {
			slaves = append(slaves, state)
		}
	}
	return slaves
}

func (db *DbState) SetMaster(config *Config) {
	config.IsMaster = true
	for _, state := range db.State {
		if state.Port == config.Port {
			state.IsMaster = true
		}
	}
}
