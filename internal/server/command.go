package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/sk25469/kv/internal/utils"
)

// Command represents a client command
type Command struct {
	Name           string // Name of the command
	CollectionName string
	Args           []string // Arguments of the command
	Result         string   // Result of the command execution
}

// ParseCommand parses a raw command string into a Command struct
func ParseCommand(rawCommand string) *Command {
	// Parse rawCommand string and extract command name and arguments
	// Split the command into parts by spaces
	parts := strings.Fields(rawCommand)
	if len(parts) == 0 {
		return nil // Ignore empty commands
	}

	// Create a new Command struct and populate its fields
	cmd := &Command{
		Name:           parts[0],
		CollectionName: parts[1],
		Args:           parts[2:],
	}
	log.Printf("current cmd: %v", cmd)
	return cmd
}

// ExecuteCommand executes a command and returns the result
func ExecuteCommand(cmd *Command, cs *CollectionStore) string {
	switch cmd.Name {
	case "SET":
		if len(cmd.Args) < 2 {
			return "Usage: SET <collection> <key> <value>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		value := strings.Join(cmd.Args[1:], " ")
		log.Printf("value is %v", value)
		cs.SetKeyInCollection(collectionName, key, value)
		return "OK"
	case "GET":
		if len(cmd.Args) < 1 {
			return "Usage: GET <collection> <key>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		return cs.GetKeyInCollection(collectionName, key)
	case "SHOWALL":
		result := cs.GetAllKeyValues()
		jsonString, err := utils.MapToJSON(result)
		if err != nil {
			log.Printf("error converting to json: %v", err)
		}
		return jsonString
	case "SHOW":
		collectionName := cmd.CollectionName
		result := cs.GetAllKeyValuesInCollection(collectionName)
		jsonString, err := utils.MapToJSON(result)
		if err != nil {
			log.Printf("error converting to json: %v", err)
		}
		return jsonString
	case "DELETE":
		if len(cmd.Args) < 1 {
			return "Usage: DELETE <collection> <key>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		cs.DeleteKeyInCollection(collectionName, key)
		return "OK"
	default:
		return fmt.Sprintf("Unknown command: %s", cmd.Name)
	}
}
