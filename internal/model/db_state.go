package models

import (
	"log"
)

type DbState struct {
	State []*Config
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
