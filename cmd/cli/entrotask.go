package main

import (
	"bufio"
	"fmt"
	"golang-secret-manager/cmd/cli/command"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// API request const
const apiRoute = "http://localhost:8080/"

// Routes
const secretUri = "secrets"
const reportUri = "reports"

// Global Vars
var userPublicKey string
var userSecretKey string
var userRegion string = "eu-north-1"
var userSavedLocation string = "./"

// Usage
const loadUsage = "Load Usage:\nload			-- loading the public + secret key from .env\nload public <key> 	-- loading public key\nload secret <key> 	-- loading secret key\nload region <region> 	-- loading the AWS region"
const getUsage = "Get Usage:\nget secrets		-- retriving all secret from AWS service\nget report <secret id> 	-- showing secret report"
const reportUsage = "Report Usage:\nget report <secret id> 	-- showing secret report"

// Reading the user input
func readInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

// Seperating the input to list of words
func tokenizeInput(input string) []string {
	return strings.Split(input, " ")
}

// Clearing the cli screen
func handleClear() {
	// clearing the console from all the text
	fmt.Print("\033[H\033[2J")
}

// Load function will load the public and secret key from the .env or manualy
func handleLoad(args []string) {
	length := len(args)
	if length == 0 || (length != 3 && length != 1) {
		fmt.Println(loadUsage)
		return
	}

	if length == 1 {
		// loading the .env file
		fmt.Println(" ---- Loading keys from the .env file ---- ")
		err := godotenv.Load(".env")
		if err != nil {
			fmt.Println("Error loading .env file: ", err)
			return
		}

		userPublicKey = os.Getenv("public")
		userSecretKey = os.Getenv("secret")

		if userPublicKey == "" || userSecretKey == "" {
			fmt.Println("Error: couldn't find public or secret key in the .env file")
			return
		}

		fmt.Println(" ---- Public key set to: '" + userPublicKey + "' ---- ")
		fmt.Println(" ---- Secret key set to: '" + userSecretKey + "' ---- ")
		return

	} else if length == 3 {
		// load public <key>
		if args[1] == "public" {
			userPublicKey = args[2]
			fmt.Println(" ---- Public key set to: '" + userPublicKey + "' ---- ")
			return
		} else if args[1] == "secret" {
			userSecretKey = args[2]
			fmt.Println(" ---- Secret key set to: '" + userSecretKey + "' ---- ")
			return
		} else if args[1] == "region" {
			// setting default zone
			userRegion = args[2]
			fmt.Println(" ---- Region set to: '" + userRegion + "' ---- ")
			return
		} else {
			fmt.Println(loadUsage)
		}
	}
}

func handleGet(args []string) {
	length := len(args)
	if length != 2 && length != 3 {
		fmt.Println(getUsage)
		return
	}

	if userPublicKey == "" || userSecretKey == "" {
		fmt.Println("Please load public and secret keys first")
		return
	}

	switch args[1] {
	case "secrets":
		fmt.Println(" ---- Getting all secrets from the server ---- ")
		com1 := command.CreateGetSecretsCommand(userPublicKey, userSecretKey, apiRoute+secretUri, userRegion)
		err := com1.Execute()
		if err != nil {
			// failed to retrive the secrets
			fmt.Println(" ----------- FAILED TO RETRIVE ----------- ")
			return
		}

		// success
		fmt.Printf(" ---- Saving all secrets to CSV file at %s ---- \n", userSavedLocation)
		com2 := command.CreateSaveToFileSecretsCommand(userSavedLocation, com1.Response)
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
			fmt.Println(reportUsage)
		}
		return
	default:
		{
			fmt.Println(getUsage)
		}
	}

}

func printBanner() {
	fmt.Println(`
    _______    _______    _______    _______    _______   _________       _______    _______    _          _______    _______    _______    _______       
   (  ____ \  (  ____ \  (  ____ \  (  ____ )  (  ____ \  \__   __/      (       )  (  ___  )  ( (    /|  (  ___  )  (  ____ \  (  ____ \  (  ____ )
   | (    \/  | (    \/  | (    \/  | (    )|  | (    \/     ) (         | () () |  | (   ) |  |  \  ( |  | (   ) |  | (    \/  | (    \/  | (    )|
   | (_____   | (__      | |        | (____)|  | (__         | |         | || || |  | (___) |  |   \ | |  | (___) |  | |        | (__      | (____)|
   (_____  )  |  __)     | |        |     __)  |  __)        | |         | |(_)| |  |  ___  |  | (\ \) |  |  ___  |  | | ____   |  __)     |     __)
	 ) |  | (        | |        | (\ (     | (           | |         | |   | |  | (   ) |  | | \   |  | (   ) |  | | \_  )  | (        | (\ (   
   /\____) |  | (____/\  | (____/\  | ) \ \__  | (____/\     | |         | )   ( |  | )   ( |  | )  \  |  | )   ( |  | (___) |  | (____/\  | ) \ \__
   \_______)  (_______/  (_______/  |/   \__/  (_______/     )_(         |/     \|  |/     \|  |/    )_)  |/     \|  (_______)  (_______/  |/   \__/
																																						  
   `)
}

func handleHelp() {
	fmt.Print("Cli Usage:\n\n")
	fmt.Println(loadUsage)
	fmt.Println()
	fmt.Println(getUsage)
}

func startCli() {
	printBanner()
	handleHelp()
	for {
		fmt.Print(">> ")
		input := readInput()
		tokens := tokenizeInput(input)

		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "help":
			handleHelp()
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
			// TODO handle clean exit
			return
		}

		fmt.Println("Invalid command for help type 'help'")
	}
}

func main() {
	startCli()
	os.Exit(0)
}
