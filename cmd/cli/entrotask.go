package main

import (
	"bufio"
	"fmt"
	"golang-secret-manager/cmd/cli/command"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const apiRoute = "http://localhost:8080/"
const secretUri = "secrets"
const reportUri = "reports"

var tempPublicKey string
var tempSecretKey string
var tempRegion string = "eu-north-1"
var tempSavedLocation string = "./"

func readInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func tokenizeInput(input string) []string {
	return strings.Split(input, " ")
}

func handleClear() {
	// clearing the console from all the text
	fmt.Print("\033[H\033[2J")
}

func handleLoad(args []string) {
	length := len(args)
	if length == 0 {
		// fmt.Println(loadUsage)
		return
	}

	if length == 3 {
		// load public <key>
		if args[1] == "public" {
			tempPublicKey = args[2]
			fmt.Println(" ---- Public key set to: '" + tempPublicKey + "' ---- ")
			return
		} else if args[1] == "secret" {
			tempSecretKey = args[2]
			fmt.Println(" ---- Secret key set to: '" + tempSecretKey + "' ---- ")
			return
		} else if args[1] == "region" {
			// setting default zone
			tempRegion = args[2]
			fmt.Println(" ---- Region set to: '" + tempRegion + "' ---- ")
			return
		} else {
			// fmt.Println(loadUsage)
			return
		}
	}

	// loading the .env file
	fmt.Println(" ---- Loading keys from the .env file ---- ")
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file: ", err)
		return
	}

	tempPublicKey = os.Getenv("public")
	tempSecretKey = os.Getenv("secret")

	if tempPublicKey == "" || tempSecretKey == "" {
		fmt.Println("Error: couldn't find public or secret key in the .env file")
		return
	}

	fmt.Println(" ---- Public key set to: '" + tempPublicKey + "' ---- ")
	fmt.Println(" ---- Secret key set to: '" + tempSecretKey + "' ---- ")
}

func handleGet(args []string) {
	if len(args) != 2 {
		// fmt.Println(getUsage)
		return
	}

	if tempPublicKey == "" || tempSecretKey == "" {
		fmt.Println("Please load public and secret keys first")
		return
	}

	switch args[1] {
	case "secrets":
		fmt.Println(" ---- Getting all secrets from the server ---- ")

		com1 := command.CreateGetSecretsCommand(tempPublicKey, tempSecretKey, apiRoute+secretUri, tempRegion)

		err := com1.Execute()
		if err != nil {
			// failed to retrive the secrets
			fmt.Println(" ----------- FAILED TO RETRIVE ----------- ")
			return
		}

		// success
		com2 := command.CreateSaveToFileSecretsCommand(tempSavedLocation, com1.Response)

		err = com2.Execute()
		if err != nil {
			// failed to retrive the secrets
			fmt.Println(" ------------- FAILED TO SAVE ------------- ")
			return
		}

		// success
		fmt.Println(" ------------- Done! ------------- ")
		return
	case "report":
		if len(args) == 3 {
			// getting report about secret args[2]
			fmt.Println(" ---- Getting report about secret '" + args[2] + "' from the server ---- ")
			// TODO
		} else {
			fmt.Println("usage: get report <secret_name> <force flag (not must)>")
		}
		return
	}

}

func startCli() {
	// printStartMenu()
	for {
		fmt.Print(">> ")
		input := readInput()
		tokens := tokenizeInput(input)

		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "help":
			// printStartMenu()
			continue
		case "load":
			handleLoad(tokens)
			continue
		case "get":
			handleGet(tokens)
			continue
		case "clear":
			handleClear()
			continue
		case "exit":
			return
		}

		fmt.Println("Invalid command for help type 'help'")
	}
}

func main() {
	startCli()
	os.Exit(0)
}
