package main

import(
	"fmt"
	"os"
	"log"
	"flag"

	"alt-llava/src/altLlava"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

var OLLAMA_PROTOCOL string
var OLLAMA_HOSTNAME string
var OLLAMA_PORT string

var OLLAMA_MODEL string = "llava"
var SILENT_OUTPUT bool = false

var writeOutputFlag string
var silentOutputFlag bool

func main(){

	loadEnvErr := godotenv.Load()
	if loadEnvErr != nil {
		log.Println(color.YellowString("Could not find .env file. Continuing with system environment variables."))
	}

	OLLAMA_PROTOCOL = os.Getenv("OLLAMA_PROTOCOL")
	OLLAMA_HOSTNAME = os.Getenv("OLLAMA_HOSTNAME")
	OLLAMA_PORT = os.Getenv("OLLAMA_PORT")

	if os.Getenv("SILENT_OUTPUT") == "true" {
		SILENT_OUTPUT = true
	} else if os.Getenv("SILENT_OUTPUT") == "false" {
		SILENT_OUTPUT = false
	}

	flag.StringVar(&writeOutputFlag, "out", "", "Write alt-text output to file path.")
	flag.BoolVar(&silentOutputFlag, "silent", false, "Only output the alt-text to the stdout.")

	flag.Parse()

	if silentOutputFlag == true {
		SILENT_OUTPUT = true
	}

	if OLLAMA_PROTOCOL == "" {

		if SILENT_OUTPUT == false {
			log.Println( color.YellowString("OLLAMA_PROTOCOL environment variable is not set. Defaulting to HTTP...") )
		}

		OLLAMA_PROTOCOL = "http"
	}

	if OLLAMA_HOSTNAME == "" {

		if SILENT_OUTPUT == false {
			log.Println( color.YellowString(`OLLAMA_HOSTNAME environment variable is not set. Defaulting to "localhost".`) )
		}

		OLLAMA_HOSTNAME = "localhost"

	}

	if OLLAMA_PORT == "" {
		
		if SILENT_OUTPUT == false {
			log.Println( color.YellowString(`OLLAMA_PORT environment variable is not set. Defaulting to 11434.`) )
		}
		OLLAMA_PORT = "11434"

	}

	if os.Getenv("OLLAMA_MODEL") != "" {
		
		if SILENT_OUTPUT == false {
			log.Println( color.MagentaString(`Setting model with OLLAMA_MODEL environment variable to "%s"`, os.Getenv("OLLAMA_MODEL") ) )
		}

		OLLAMA_MODEL = os.Getenv("OLLAMA_MODEL")
	}

	imageToRetrieve := os.Getenv("IMAGE_URL")

	if imageToRetrieve == "" {
		log.Fatal(color.RedString(`No IMAGE_URL was passed for processing. Exiting...`))
	}

	altLlavaSettings := map[string]interface{}{
		"protocol" : OLLAMA_PROTOCOL,
		"port" : OLLAMA_PORT,
		"host" : OLLAMA_HOSTNAME,
		"model" : OLLAMA_MODEL,
		"silent" : SILENT_OUTPUT,
	}

	altLlava.Init(altLlavaSettings)

	altText, altTextErr := altLlava.GenerateAltTextForImage(imageToRetrieve)

	if altTextErr != nil {
		log.Fatal( color.RedString(`Could not process image "%s": %s`, imageToRetrieve, altTextErr.Error()) )
	}

	fmt.Println(color.GreenString(altText))

	if writeOutputFlag != ""{
		writeOutputToFileErr := altLlava.WriteAltTextToFile(altText, writeOutputFlag)

		if writeOutputToFileErr != nil{
			log.Println(color.RedString(`Could not write alt-text to file "%s"`, writeOutputFlag))
		}

	}

	if os.Getenv("PUBLISH_TARGET") != "" {

		publishTarget := os.Getenv("PUBLISH_TARGET")

		switch publishTarget{
			case "s3":

				resultsKey, pubErr := altLlava.PublishAltTextResultsToS3(altText)

				if pubErr != nil {
					log.Println( color.RedString("Could not publish alt-text results to S3: %s", pubErr.Error()) )
					return
				}

				log.Println(color.CyanString("Results written to S3 Bucket with Key: %s", resultsKey))

		}

	}

}