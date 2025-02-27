package main

import(
	"fmt"
	"os"
	"log"
	"alt-llava/ollamaInterface"

	
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

var OLLAMA_PROTOCOL string
var OLLAMA_HOST string
var OLLAMA_PORT string

var OLLAMA_ORIGIN string

func main(){

	loadEnvErr := godotenv.Load()
	if loadEnvErr != nil {
		log.Println(color.YellowString("Could not find .env file. Continuing with system environment variables."))
	}

	OLLAMA_PROTOCOL = os.Getenv("OLLAMA_PROTOCOL")
	OLLAMA_HOST = os.Getenv("OLLAMA_HOST")
	OLLAMA_PORT = os.Getenv("OLLAMA_PORT")

	if OLLAMA_PROTOCOL == "" {
		log.Println( color.YellowString("OLLAMA_PROTOCOL environment variable is not set. Defaulting to HTTP.") )
		OLLAMA_PROTOCOL = "http"
	}

	if OLLAMA_HOST == "" {
		log.Fatal( color.RedString("OLLAMA_HOST environment variable is not set. Exiting.") )
	}

	if OLLAMA_PORT == "" {
		log.Fatal( color.RedString("OLLAMA_PORT environment variable is not set. Exiting.") )
	}

	OLLAMA_ORIGIN = fmt.Sprintf("%s://%s:%s", OLLAMA_PROTOCOL, OLLAMA_HOST, OLLAMA_PORT)

	log.Println( color.CyanString(`OLLAMA_ORIGIN set as "%s"`, OLLAMA_ORIGIN) )

	_ = ollamaInterface.NewClient(OLLAMA_ORIGIN)
	
    // Example: List local models
    // localModels, err := client.ListLocalModels()

}