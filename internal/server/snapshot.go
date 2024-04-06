package server

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

// WriteCommandsToFile writes a slice of Command structs to a file
func WriteCommandsToFile(command Command, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	cmdBytes, err := json.Marshal(command)
	if err != nil {
		return err
	}
	_, err = writer.Write(cmdBytes)
	if err != nil {
		return err
	}
	_, err = writer.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

// ReadCommandsFromFile reads a slice of Command structs from a file
func ReadCommandsFromFile(filename string) ([]Command, error) {
	var commands []Command

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("no such file to open")
		return []Command{}, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var cmd Command
		err := json.Unmarshal(scanner.Bytes(), &cmd)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}
