package model

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	network_model "github.com/sk25469/kv/internal/network/model"
)

// CommandType represents the type of command
type CommandType string

const (
	Set          CommandType = "SET"
	Get          CommandType = "GET"
	Delete       CommandType = "DEL"
	IAM          CommandType = "COMM:IAM"
	HEALTH_CHECK CommandType = "COMM:HEALTH_CHECK"
	ECHO         CommandType = "COMM:ECHO"
	STOP         CommandType = "COMM:STOP"
)

type CommunicationModel struct {
	ID       uuid.UUID                 `json:"id"`
	Command  CommandType               `json:"command"`
	SendTo   *network_model.NodeConfig `json:"send_to"`
	SentFrom *network_model.NodeConfig `json:"sent_from"`
}

func (c *CommunicationModel) ToBytes() ([]byte, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return []byte{}, err
	}
	jsonData = append(jsonData, '\n')
	return jsonData, nil
}

func NewCommunicationModel(cmdType CommandType, sendTo, sentFrom *network_model.NodeConfig) *CommunicationModel {
	return &CommunicationModel{
		Command:  cmdType,
		SendTo:   sendTo,
		SentFrom: sentFrom,
		ID:       uuid.New(),
	}
}

func (c *CommunicationModel) FromJSON(data []byte) {
	json.Unmarshal(data, c)
}

func CommunicationModelToJSON(data []byte) (*CommunicationModel, error) {
	var model CommunicationModel
	err := json.Unmarshal(data, &model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (c *CommunicationModel) Decode() ([]byte, error) {
	return c.ToBytes()
}

type Command struct {
	ID    uuid.UUID
	Type  CommandType
	Name  string   // Name of the command
	Args  []string // Arguments of the command
	Key   string
	Value string
}

// ParseCommand parses a raw command string into a Command struct
func (c *Command) Encode(rawCommand string) *Command {
	// Parse rawCommand string and extract command name and arguments
	// Split the command into parts by spaces
	parts := strings.Fields(rawCommand)
	if len(parts) == 0 {
		return nil // Ignore empty commands
	}

	Args := []string{}
	if len(parts) > 1 {
		Args = parts[1:]
	}

	var key, value string

	if len(Args) > 1 {
		key = Args[len(Args)-2]
		value = Args[len(Args)-1]
	} else if len(Args) == 1 {
		key = Args[0]
	}

	// Create a new Command struct and populate its fields
	cmd := &Command{
		Name:  parts[0],
		Args:  Args,
		Key:   key,
		Value: value,
		ID:    uuid.New(),
	}

	// Determine the command type based on the command name
	switch cmd.Name {
	case "SET":
		cmd.Type = Set
	case "GET":
		cmd.Type = Get
	case "DEL":
		cmd.Type = Delete
	}

	return cmd
}

func (c *Command) Decode() []byte {
	return []byte(c.Name + " " + strings.Join(c.Args, " ") + "\n")
}
