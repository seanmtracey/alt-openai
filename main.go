package main

import(
	"fmt"
	"os"
	"log"
	"strings"
	"flag"
	"path/filepath"
	"net/http"
	"io/ioutil"
	"encoding/base64"
	
	"alt-llava/ollamaInterface"
	
	"github.com/google/uuid"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

var OLLAMA_PROTOCOL string
var OLLAMA_HOST string
var OLLAMA_PORT string

var OLLAMA_ORIGIN string
var OLLAMA_MODEL string = "llava"
var SILENT_OUTPUT bool = false

var writeOutputFlag string
var silentOutputFlag bool

func downloadImageFromURLToDir(url, dest string) (string, error) {

	outputFileUUID := uuid.New().String()

	if SILENT_OUTPUT == false {
		log.Println( color.CyanString(`Attempting to download image from "%s" and store it in "%s"`, url, dest) )
	}

	// Go and get the file from the passed URL...
	response, httpErr := http.Get(url)
    if httpErr != nil {
		log.Println( color.RedString("Failed to get file from URL: %s", httpErr.Error() ) )
		return "", httpErr
    }
    defer response.Body.Close()

	fileBytes, readBytesErr := ioutil.ReadAll(response.Body)
	if readBytesErr != nil {
		log.Println( color.RedString(`Could not read bytes for mimetype section from URL "%s": %s`, url, readBytesErr.Error() ) )
		return "", readBytesErr
	}

	mimeType := http.DetectContentType(fileBytes)
	imageExtension := strings.Split(mimeType, "/")[1]
	outputFilePath := fmt.Sprintf("%s/%s.%s", dest, outputFileUUID, imageExtension)

	// Create a file on the system to put it...
	file, fileCreationErr := os.Create(outputFilePath)
    if fileCreationErr != nil {
		log.Println( color.RedString(`Failed to get create file "%s" for storage: %s`, dest, fileCreationErr.Error() ) )
		return "", fileCreationErr
    }
    defer file.Close()

	// ... and then dump the bytes into the file 👍
	_, fileWriteErr := file.Write(fileBytes)
	if fileWriteErr != nil {
		log.Println( color.RedString(`Could not write HTTP response body (%s) to file "%s": %s`, url, dest, fileWriteErr.Error() ) )
		return "", fileWriteErr
	}

	return outputFilePath, nil

}

func convertImageToBase64(imagePath string) (string, error) {
    // Read the file into a []byte
    fileData, err := os.ReadFile(imagePath)
    if err != nil {
        return "", err
    }

    // Encode the bytes to a base64 string
    encoded := base64.StdEncoding.EncodeToString(fileData)
    return encoded, nil
}

func generateAltTextForImage(imageURL string) (string, error) {

	imagePath, downloadErr := downloadImageFromURLToDir(imageURL, "./images")

	if downloadErr != nil {
		log.Println( color.RedString(`Could not download "%s" image for processing: %s`, imageURL, downloadErr.Error()) )
		return "", downloadErr
	}

	base64, b64Err := convertImageToBase64(imagePath)

	if b64Err != nil {
		log.Println( color.RedString(`Failed to convert image "%s" to Base 64 encoding: %s`, imagePath, b64Err.Error()) )
		return "", b64Err
	}

	if SILENT_OUTPUT == false {
		log.Println(color.MagentaString("Base64 image data: %s...", base64[:50]))
	}

	ollama := ollamaInterface.NewClient(OLLAMA_ORIGIN)

	prompt := map[string]interface{}{
		"model" : OLLAMA_MODEL,
		"prompt" : "Briefly, what is in this image?",
		"images" : []string{
			base64,
		},
	}

	if SILENT_OUTPUT == false {
		log.Println(color.CyanString(`Running prompt with model "%s"`, OLLAMA_MODEL))
		log.Println(color.CyanString(`Prompt for image: "%s"`, prompt["prompt"].(string)))
	}

	llavaResponse, llavaErr := ollama.GenerateCompletion(prompt)

	if llavaErr != nil {
		log.Println( color.RedString("Failed to get completion from Ollama: %s", llavaErr) )
		return "", nil
	}

	responseString := ""

	for _, part := range llavaResponse {
		responseString += part["response"].(string)
	}

	return strings.TrimSpace(responseString), nil

}

func writeAltTextToFile(altText, outputPath string) error {
    // Optionally create parent directories if they don't exist.
    dir := filepath.Dir(outputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
		log.Println( color.RedString("Failed to make directory to write output file: %s", err.Error()) )
        return fmt.Errorf("failed to create directory %s: %w", dir, err)
    }

    file, err := os.OpenFile(outputPath, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)
    if err != nil {
		log.Println( color.RedString("Failed to open output file to write results: %s", err.Error()) )
        return fmt.Errorf("failed to open file %s: %w", outputPath, err)
    }

    defer file.Close()

    // Write the altText to the file
    _, err = file.WriteString(altText)
    if err != nil {
		log.Println( color.RedString("Failed to write output to file: %s", err.Error()) )
        return fmt.Errorf("failed to write alt text to %s: %w", outputPath, err)
    }

    return nil
}

func main(){

	loadEnvErr := godotenv.Load()
	if loadEnvErr != nil {
		log.Println(color.YellowString("Could not find .env file. Continuing with system environment variables."))
	}

	OLLAMA_PROTOCOL = os.Getenv("OLLAMA_PROTOCOL")
	OLLAMA_HOST = os.Getenv("OLLAMA_HOST")
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

	if OLLAMA_HOST == "" {

		if SILENT_OUTPUT == false {
			log.Println( color.YellowString(`OLLAMA_HOST environment variable is not set. Defaulting to "localhost".`) )
		}

		OLLAMA_HOST = "localhost"

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

	OLLAMA_ORIGIN = fmt.Sprintf("%s://%s:%s", OLLAMA_PROTOCOL, OLLAMA_HOST, OLLAMA_PORT)

	if SILENT_OUTPUT == false {
		log.Println( color.CyanString(`OLLAMA_ORIGIN set as "%s"`, OLLAMA_ORIGIN) )
	}

	imageToRetrieve := os.Getenv("IMAGE_URL")

	if imageToRetrieve == "" {
		log.Fatal(color.RedString(`No IMAGE_URL was passed for processing. Exiting...`))
	}

	altText, altTextErr := generateAltTextForImage(imageToRetrieve)

	if altTextErr != nil {
		log.Fatal( color.RedString(`Could not process image "%s": %s`, imageToRetrieve, altTextErr.Error()) )
	}

	fmt.Println(color.GreenString(altText))

	if writeOutputFlag != ""{
		writeOutputToFileErr := writeAltTextToFile(altText, writeOutputFlag)

		if writeOutputToFileErr != nil{
			log.Println(color.RedString(`Could not write alt-text to file "%s"`, writeOutputFlag))
		}

	}

}