package command

import (
	"fmt"
	"golang-secret-manager/types"
	"os"
)

type SaveToFileSecretsCommand struct {
	PathToSave string
	value      types.GetAllSecretsResponse
}

func CreateSaveToFileSecretsCommand(PathToSave string, val types.GetAllSecretsResponse) *SaveToFileSecretsCommand {
	return &SaveToFileSecretsCommand{
		PathToSave: PathToSave,
		value:      val,
	}
}

func (s *SaveToFileSecretsCommand) Execute() error {
	file, err := os.Create(s.PathToSave + "/secrets.csv")
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()
	for _, secret := range s.value.Secrets {
		_, err = file.WriteString(fmt.Sprintf("%s,%s\n%s,%s\n", "Name", "ARN", secret.Name, secret.ARN))
		if err != nil {
			return fmt.Errorf("error writing to CSV: %v", err)
		}
		temp := s.value.AccessLog[secret.ARN]
		length := len(temp)
		if length > 0 {
			_, err = file.WriteString("ACCESS LOG\nUser, Event Time, Event Name\n")
			if err != nil {
				return fmt.Errorf("error writing to CSV: %v", err)
			}

			for _, accessLog := range temp {
				_, err = file.WriteString(fmt.Sprintf("%s,%s,%s\n", accessLog.User, accessLog.EventTime, accessLog.EventName))
				if err != nil {
					return fmt.Errorf("error writing to CSV: %v", err)
				}
			}
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing to CSV: %v", err)
		}
	}
	return nil
}
