package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// GenerateBase64ClientID generates a Base64 client ID based on the client's IP address
func GenerateBase64ClientID(ipAddress string) (string, error) {
	// Parse the IP address string
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	// Hash the IP address using SHA256
	hashed := sha256.Sum256(ip)

	// Encode the hashed value in Base64
	base64ID := base64.StdEncoding.EncodeToString(hashed[:])

	return base64ID, nil
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
