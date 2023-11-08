package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang-secret-manager/types"
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
		fmt.Println("Error retrieving secrets from server: ", err)
		return err
	}

	defer req.Body.Close()
	var res types.GetAllSecretsResponse
	if req.StatusCode == 200 {
		// retriving the list of secrets from the request
		err := json.NewDecoder(req.Body).Decode(&res)
		if err != nil {
			fmt.Println("Error decoding response: ", err)
			return err
		}
		s.Response = res
	} else {
		// printing the error
		var Response struct {
			Error string `json:"error"`
		}
		json.NewDecoder(req.Body).Decode(&Response)
		fmt.Println("Error retrieving secrets from server: ", Response.Error)
		return fmt.Errorf(res.Error)
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
		fmt.Println("Error retrieving secrets from server")
		return err
	}

	defer req.Body.Close()

	var res types.GetReportResponse
	if req.StatusCode == 200 {
		// retriving the list of secrets from the request
		err := json.NewDecoder(req.Body).Decode(&res)
		if err != nil {
			fmt.Println("Error decoding response: ", err)
			return err
		}
		s.Response = res
	} else {
		// printing the error
		res.Error = "Failed to retrive report from the server"
		json.NewDecoder(req.Body).Decode(&res)
		fmt.Println("Error retrieving report for", s.SecretID, "from server!")
		return fmt.Errorf(res.Error)
	}
	return nil
}
