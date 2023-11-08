package command

import (
	"bytes"
	"fmt"
	"golang-secret-manager/types"
	GenericEncoding "golang-secret-manager/utils/genericEncoding"
	"net/http"
)

type GetSecretsCommand struct {
	PublicKey string
	SecretKey string
	ApiRoute  string
	Region    string
	Response  types.GetAllSecretsResponse
}

func CreateGetSecretsCommand(PublicKey string,
	SecretKey string,
	ApiRoute string,
	Region string) *GetSecretsCommand {
	return &GetSecretsCommand{
		PublicKey: PublicKey,
		SecretKey: SecretKey,
		ApiRoute:  ApiRoute,
		Region:    Region,
	}
}

func (s *GetSecretsCommand) Execute() error {

	// Request Body
	// PublicKey string `json:"public_key"`
	// SecretKey string `json:"secret_key"`
	// Region    string `json:"region"`

	data := []byte(fmt.Sprintf(`{"public_key": "%s", "secret_key": "%s",  "region":"%s"}`, s.PublicKey, s.SecretKey, s.Region))
	payload := bytes.NewBuffer(data)

	// sending to the server using POST request
	req, err := http.Post(s.ApiRoute, "application/json", payload)
	if err != nil {
		// Handle the error
		return fmt.Errorf("error retrieving secrets from server: %v", err)
	}

	defer req.Body.Close()
	if req.StatusCode == http.StatusOK {
		// retriving the list of secrets from the request
		valRes, err := GenericEncoding.JsonBodyDecoder[types.GetAllSecretsResponse](req.Body)
		if err != nil {
			return fmt.Errorf("failed to decode error from the server: %v", err)
		}
		s.Response = *valRes
	} else {
		// printing the error
		valErr, err := GenericEncoding.JsonBodyDecoder[types.ApiError](req.Body)
		if err != nil {
			// failed to decode error
			return fmt.Errorf("failed to decode error from the server")
		} else {
			return fmt.Errorf(valErr.Err)
		}
	}
	return nil
}

type GetReportByIdCommand struct {
	PublicKey string
	SecretKey string
	SecretID  string
	ApiRoute  string
	Region    string
	Response  types.GetReportResponse
}

func CreateGetReportByIdCommand(PublicKey string,
	SecretKey string,
	SecretID string,
	ApiRoute string,
	Region string) *GetReportByIdCommand {
	return &GetReportByIdCommand{
		PublicKey: PublicKey,
		SecretKey: SecretKey,
		SecretID:  SecretID,
		ApiRoute:  ApiRoute,
		Region:    Region,
	}
}

func (s *GetReportByIdCommand) Execute() error {

	// Request Body
	// PublicKey string `json:"public_key"`
	// SecretKey string `json:"secret_key"`
	// Region    string `json:"region"`
	// SecretID  string `json:"secret_id"`

	data := []byte(fmt.Sprintf(`{"public_key": "%s", "secret_key": "%s",  "region":"%s", "secret_id":"%s"}`, s.PublicKey, s.SecretKey, s.Region, s.SecretID))
	payload := bytes.NewBuffer(data)

	// sending to the server using POST request
	req, err := http.Post(s.ApiRoute, "application/json", payload)
	if err != nil {
		// Handle the error
		return fmt.Errorf("error retrieving secrets from server: %v", err)
	}

	defer req.Body.Close()

	if req.StatusCode == http.StatusOK {
		// retriving the list of secrets from the request
		valRes, err := GenericEncoding.JsonBodyDecoder[types.GetReportResponse](req.Body)
		if err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
		s.Response = *valRes
	} else {
		// printing the error
		valErr, err := GenericEncoding.JsonBodyDecoder[types.ApiError](req.Body)
		if err != nil {
			// failed to decode error
			return fmt.Errorf("failed to retrive report from the server")
		} else {
			return fmt.Errorf(valErr.Err)
		}
	}
	return nil
}
