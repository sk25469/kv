package utils

import "encoding/json"

// MapToString converts a map[string]map[string]string to a string
func MapToJSON(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
