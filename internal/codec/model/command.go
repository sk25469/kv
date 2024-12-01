package model

import "strings"

// CommandType represents the type of command
type CommandType string

const (
	Set    CommandType = "SET"
	Get    CommandType = "GET"
	Delete CommandType = "DEL"
)

type Command struct {
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

func (c *Command) Decode() string {
	return c.Name + " " + strings.Join(c.Args, " ")
}
