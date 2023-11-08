package GenericEncoding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HelperFunc to write back json to the client
func WriteJson(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}

// Decoding the json body from a request
func JsonBodyDecoder[T any](r io.Reader) (*T, error) {
	var v T
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ToJson serializes the input value to JSON
func ToJson[T any](value T) ([]byte, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error converting to JSON: %v", err)
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
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}
	return &valueToReturn, nil
}
