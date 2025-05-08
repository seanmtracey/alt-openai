package main

import(
	"fmt"
	"os"
	"log"
	"flag"

	"alt-openai/src/altOpenAI"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

var OPEN_AI_ORIGIN string
var OPEN_AI_KEY string
var OPEN_AI_MODEL string = "gpt-4-turbo"
var SILENT_OUTPUT bool = false
var S3_KEY_ONLY bool = false

var writeOutputFlag string
var silentOutputFlag bool

func main(){

	loadEnvErr := godotenv.Load()
	if loadEnvErr != nil {

		if os.Getenv("SILENT_OUTPUT") != "true"{

			log.Println(color.YellowString("Could not find .env file. Continuing with system environment variables."))

		}

	}

	OPEN_AI_ORIGIN = os.Getenv("OPEN_AI_ORIGIN")
	OPEN_AI_KEY = os.Getenv("OPEN_AI_KEY")

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

	if OPEN_AI_ORIGIN == "" {

		if SILENT_OUTPUT == false {
			log.Println( color.YellowString(`OPEN_AI_ORIGIN environment variable is not set. Defaulting to "api.openai.com".`) )
		}

		OPEN_AI_ORIGIN = "api.openai.com"

	}

	if OPEN_AI_ORIGIN == "" {

		log.Println( color.RedString(`OPEN_AI_KEY environment variable is not set. Exiting...`) )

		os.Exit(1)

	}

	if os.Getenv("OPEN_AI_MODEL") != "" {
		
		if SILENT_OUTPUT == false {
			log.Println( color.MagentaString(`Setting model with OPEN_AI_MODEL environment variable to "%s"`, os.Getenv("OPEN_AI_MODEL") ) )
		}

		OPEN_AI_MODEL = os.Getenv("OPEN_AI_MODEL")

	}

	if os.Getenv("S3_KEY_ONLY") == "true" {
		S3_KEY_ONLY = true
	}

	imageToRetrieve := os.Getenv("IMAGE_URL")

	if imageToRetrieve == "" {
		log.Fatal(color.RedString(`No IMAGE_URL was passed for processing. Exiting...`))
	}

	openAISettings := map[string]interface{}{
		"origin" : OPEN_AI_ORIGIN,
		"model" : OPEN_AI_MODEL,
		"api_key" : OPEN_AI_KEY,
		"silent" : SILENT_OUTPUT,
	}

	altOpenAI.Init(openAISettings)

	altText, altTextErr := altOpenAI.GenerateAltTextForImage(imageToRetrieve)

	if altTextErr != nil {
		log.Fatal( color.RedString(`Could not process image "%s": %s`, imageToRetrieve, altTextErr.Error()) )
	}

	if writeOutputFlag != ""{
		writeOutputToFileErr := altOpenAI.WriteAltTextToFile(altText, writeOutputFlag)

		if writeOutputToFileErr != nil{
			log.Println(color.RedString(`Could not write alt-text to file "%s"`, writeOutputFlag))
		}

	}

	if os.Getenv("PUBLISH_TARGET") != "" {

		publishTarget := os.Getenv("PUBLISH_TARGET")

		switch publishTarget{
			case "s3":

				resultsKey, pubErr := altOpenAI.PublishAltTextResultsToS3(altText)

				if pubErr != nil {
					log.Println( color.RedString("Could not publish alt-text results to S3: %s", pubErr.Error()) )
					return
				}

				if S3_KEY_ONLY == true {
					fmt.Println(resultsKey)
				}

		}

	}

	if S3_KEY_ONLY != true {
		fmt.Println(color.GreenString(altText))
	}

}