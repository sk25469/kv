package server

import (
	"fmt"
	"log"
	"strings"

	models "github.com/sk25469/kv/internal/model"
	"github.com/sk25469/kv/utils"
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

	collectionName := ""
	Args := []string{}
	if len(parts) >= 2 {
		collectionName = parts[1]
	}
	if len(parts) > 2 {
		Args = parts[2:]
	}

	// Create a new Command struct and populate its fields
	cmd := &Command{
		Name:           parts[0],
		CollectionName: collectionName,
		Args:           Args,
	}
	log.Printf("current cmd: %v", cmd)
	return cmd
}

// ExecuteCommand executes a command and returns the result
func ExecuteCommand(cmd *Command, cs *models.CollectionStore, ts *models.TransactionalKeyValueStore, cc *models.ClientConfig, kv *models.KVServer, ps *models.PubSub) string {
	switch cmd.Name {
	case "AUTH":
		if !kv.Config.ProtectedMode {
			return "No need of password without protected mode"
		}
		if len(cmd.Args) < 1 {
			return "Usage: AUTH <username> <password>"
		}
		username := cmd.CollectionName
		password := cmd.Args[0]
		result, ok := kv.Authenticate(username, password)
		if !ok {
			return result
		}
		cc.ClientState.IsAuthenticated = true
		return result
	case "BEGIN":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		ts.BeginTransaction()
		cc.ClientState.State = utils.TRANSACTIONAL
		return "OK"
	case "COMMIT":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		ts.ExecTransaction()
		cc.ClientState.State = utils.ACTIVE
		return "OK"
	case "ROLLBACK":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		ts.RollbackTransaction()
		cc.ClientState.State = utils.ACTIVE
		return "OK"
	case "TSET":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		if cc.ClientState.State != utils.TRANSACTIONAL {
			return "ERROR: Transaction not started"
		}
		if len(cmd.Args) < 2 {
			return "Usage: TSET <collection_name> <key> <value>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		value := strings.Join(cmd.Args[1:], " ")
		ts.Set(collectionName, key, value)
		return "OK"
	case "TGET":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		if cc.ClientState.State != utils.TRANSACTIONAL {
			return "ERROR: Transaction not started"
		}
		if len(cmd.Args) < 1 {
			return "Usage: TSET <collection_name> <key> <value>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		val, err := ts.Get(collectionName, key)
		if err != nil {
			log.Printf("error getting key from transaction: %v", err)
			return "ERROR"
		}
		return val
	case "SET-TTL":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		if len(cmd.Args) < 1 {
			return "Usage: SET-TTL <collection> <key> <ttl>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		ttl := cmd.Args[1]
		duration, err := utils.ParseDuration(ttl)
		if err != nil {
			log.Printf("invalid time format: %v", err)
			return "Usage: SET-TTL <collection> <key> <ttl (xm xhxm xxs)>"
		}
		cs.UpdateKeyInCollectionWithTTL(collectionName, key, duration)
		return "OK"
	case "SET":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		if len(cmd.Args) < 2 {
			return "Usage: SET <collection> <key> <value>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		value := strings.Join(cmd.Args[1:], " ")
		cs.SetKeyInCollection(collectionName, key, value)
		return "OK"
	case "GET":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		if len(cmd.Args) < 1 {
			return "Usage: GET <collection> <key>"
		}
		key := cmd.Args[0]
		collectionName := cmd.CollectionName
		return cs.GetKeyInCollection(collectionName, key)
	case "SHOWALL":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		result := cs.GetAllKeyValues()
		jsonString, err := utils.MapToJSON(result)
		if err != nil {
			log.Printf("error converting to json: %v", err)
		}
		return jsonString
	case "SHOW":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
		collectionName := cmd.CollectionName
		result := cs.GetAllKeyValuesInCollection(collectionName)
		jsonString, err := utils.MapToJSON(result)
		if err != nil {
			log.Printf("error converting to json: %v", err)
		}
		return jsonString
	case "DELETE":
		if !cc.ClientState.IsAuthenticated && kv.Config.ProtectedMode {
			return "unauthorized"
		}
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

func ShouldWriteLog(cmd Command) bool {
	if cmd.Name == utils.SET || cmd.Name == utils.DEL || cmd.Name == utils.SET_TTL || cmd.Name == utils.SUBSCRIBE || cmd.Name == utils.PUBLISH {
		return true
	}
	return false
}
