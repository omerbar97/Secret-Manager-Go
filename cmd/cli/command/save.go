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
		fmt.Println("Error creating CSV file: ", err)
		return err
	}
	defer file.Close()
	for _, secret := range s.value.Secrets {
		_, err = file.WriteString(fmt.Sprintf("%s,%s\n%s,%s\n", "Name", "ARN", secret.Name, secret.ARN))
		if err != nil {
			fmt.Println("Error writing to CSV: ", err)
			return err
		}
		temp := s.value.AccessLog[secret.ARN]
		length := len(temp)
		if length > 0 {
			_, err = file.WriteString("ACCESS LOG\nUser, Event Time, Event Name\n")
			if err != nil {
				fmt.Println("Error writing to CSV: ", err)
				return err
			}

			for _, accessLog := range temp {
				_, err = file.WriteString(fmt.Sprintf("%s,%s,%s\n", accessLog.User, accessLog.EventTime, accessLog.EventName))
				if err != nil {
					fmt.Println("Error writing to CSV: ", err)
					return err
				}
			}
		}
		file.WriteString("\n")
	}
	return nil
}
