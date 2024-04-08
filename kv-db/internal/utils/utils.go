package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GetParsedIP(ip string) string {
	return strings.Split(ip, ":")[0]
}

// GenerateBase64ClientID generates a Base64 client ID based on the client's IP address
func GenerateBase64ClientID(ipAddress string) (string, error) {
	// Parse the IP address string
	// Generate a random UUID
	uuid := uuid.New()

	// Convert UUID to string
	uuidStr := uuid.String()

	return uuidStr, nil
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
	parts := strings.Split(input, "")
	duration := time.Duration(0)

	for i := 0; i < len(parts); i += 2 {
		value, err := strconv.Atoi(parts[i])
		if err != nil {
			return 0, err
		}

		unit := parts[i+1]
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
