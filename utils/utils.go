package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// a random no. between 1 and 100
func GetShardID() int {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 1 and 100
	randomNumber := rand.Intn(100) + 1
	return randomNumber
}

func ContainsPubSub(cmd string) bool {
	return strings.Contains(cmd, "SUBSCRIBE") || strings.Contains(cmd, "PUBLISH")
}

func GetParsedIP(ip string) string {
	return strings.Split(ip, ":")[0]
}

// GenerateBase64ClientID generates a Base64 client ID based on the client's IP address
func GenerateBase64ClientID() string {
	// Parse the IP address string
	// Generate a random UUID
	uuid := uuid.New()

	// Convert UUID to string
	uuidStr := uuid.String()

	return uuidStr
}

func GetCurrentTime() time.Time {
	return time.Now()
}

// MapToString converts a map[string]map[string]string to a string
func MapToJSON(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
func ParseDuration(input string) (time.Duration, error) {
	// Regular expression to match durations like "10s", "5m", "2h", etc.
	re := regexp.MustCompile(`(\d+)([hms])`)
	matches := re.FindAllStringSubmatch(input, -1)

	if matches == nil {
		return 0, fmt.Errorf("invalid duration format")
	}

	var duration time.Duration
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}

		unit := match[2]
		log.Printf("value: %v --- unit %v", value, unit)
		switch unit {
		case "h":
			duration += time.Duration(value) * time.Hour
		case "m":
			duration += time.Duration(value) * time.Minute
		case "s":
			duration += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("invalid unit: %s", unit)
		}
	}

	return duration, nil
}

func CreateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error generating hashed password: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func ParseMaxConnections(maxConnStr string) int {
	maxConn, err := strconv.Atoi(maxConnStr)
	if err != nil {
		// Handle error
		log.Printf("unable to parse maxConn: %v", err)
		return 10
	}
	return maxConn
}

func ConvertStructToJSON(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func AsciiArt() {
	art := `          _____               _____          
         /\    \             /\    \         
        /::\____\           /::\____\        
       /:::/    /          /:::/    /        
      /:::/    /          /:::/    /         
     /:::/    /          /:::/    /          
    /:::/____/          /:::/____/           
   /::::\    \          |::|    |            
  /::::::\____\________ |::|    |     _____  
 /:::/\:::::::::::\    \|::|    |    /\    \ 
/:::/  |:::::::::::\____|::|    |   /::\____\
\::/   |::|~~~|~~~~~    |::|    |  /:::/    /
 \/____|::|   |         |::|    | /:::/    / 
       |::|   |         |::|____|/:::/    /  
       |::|   |         |:::::::::::/    /   
       |::|   |         \::::::::::/____/    
       |::|   |          ~~~~~~~~~~          
       |::|   |                              
       \::|   |                              
        \:|   |                              
         \|___|                              
                                             
			`
	fmt.Println(art)
}
