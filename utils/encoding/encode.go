package GenericEncoding

import (
	"encoding/json"
	"fmt"
)

// ToJson serializes the input value to JSON
func ToJson[T any](value T) ([]byte, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("Error converting to JSON: %v", err)
	}
	return jsonData, nil
}

// FromJson deserializes the JSON data into the input value
func FromJson[T any](data interface{}) (*T, error) {
	var valueToReturn T
	bytesVal, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("couldn't convert type interface{} to []byte")
	}
	err := json.Unmarshal(bytesVal, &valueToReturn)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}
	return &valueToReturn, nil
}
