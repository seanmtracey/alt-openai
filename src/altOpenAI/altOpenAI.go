package altOpenAI

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"alt-openai/src/s3uploader"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

var OPEN_AI_ORIGIN string
var OPEN_AI_URL string
var OPEN_AI_MODEL string = "gpt-4-turbo"
var OPEN_AI_KEY string
var SILENT_OUTPUT bool = false

func Init(settings map[string]interface{}) {

	if settings["model"] != nil {
		OPEN_AI_MODEL = settings["model"].(string)
	}

	if settings["silent"] != nil {
		SILENT_OUTPUT = settings["silent"].(bool)
	}

	if settings["origin"] != nil {
		OPEN_AI_ORIGIN = settings["origin"].(string)
	}

	if settings["api_key"] != nil {
		OPEN_AI_KEY = settings["api_key"].(string)
	}

	OPEN_AI_URL = fmt.Sprintf("https://%s/v1/chat/completions", OPEN_AI_ORIGIN)

	if !SILENT_OUTPUT {
		log.Println(color.CyanString(`OPEN_AI_ORIGIN set as "%s"`, OPEN_AI_ORIGIN))
	}

}

func downloadImageFromURL(url string) ([]byte, error) {

	if !SILENT_OUTPUT {
		log.Println(color.CyanString(`Attempting to download image from "%s"`, url))
	}

	response, err := http.Get(url)
	if err != nil {
		log.Println(color.RedString("Failed to get file from URL: %s", err.Error()))
		return nil, err
	}
	defer response.Body.Close()

	fileBytes, readBytesErr := io.ReadAll(response.Body)
	if readBytesErr != nil {
		log.Println(color.RedString(`Could not read bytes from URL "%s": %s`, url, readBytesErr.Error()))
		return nil, readBytesErr
	}

	return fileBytes, nil

}

func convertImageBytesToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}

func GenerateAltTextForImage(imageURL string) (string, error) {

	PROMPT_TEXT := os.Getenv("PROMPT_TEXT")
	if PROMPT_TEXT == "" {
		PROMPT_TEXT = "Briefly, what is in this image?"
	}

	imgData, downloadErr := downloadImageFromURL(imageURL)
	if downloadErr != nil {
		log.Println(color.RedString(`Could not download "%s" image for processing: %s`, imageURL, downloadErr.Error()))
		return "", downloadErr
	}

	base64Image := convertImageBytesToBase64(imgData)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	if !SILENT_OUTPUT {
		log.Println(color.MagentaString("Base64 image data: %s...", dataURL[:50]))
	}

	reqBody := map[string]interface{}{
		"model": OPEN_AI_MODEL,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": PROMPT_TEXT},
					{"type": "image_url", "image_url": map[string]string{
						"url": dataURL,
					}},
				},
			},
		},
		"max_tokens": 300,
	}

	if !SILENT_OUTPUT {
		log.Println(color.CyanString(`Running prompt with model "%s"`, OPEN_AI_MODEL))
		log.Println(color.CyanString(`Prompt for image: "%s"`, PROMPT_TEXT))
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(color.RedString("Failed to marshal request body: %s", err.Error()))
		return "", err
	}

	req, reqErr := http.NewRequest("POST", OPEN_AI_URL, bytes.NewBuffer(bodyBytes))
	if reqErr != nil {
		log.Println(color.RedString(`Could not make request to OpenAI "%s" to process image: %s`, OPEN_AI_URL, reqErr.Error()))
		return "", reqErr
	}

	req.Header.Set("Authorization", "Bearer "+OPEN_AI_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(color.RedString("HTTP request failed: %s", err.Error()))
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(color.RedString("Failed to read response body: %s", err.Error()))
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(color.RedString("OpenAI API returned status code %d: %s", resp.StatusCode, string(respBody)))
		return "", fmt.Errorf("OpenAI API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Println(color.RedString("Failed to unmarshal response: %s", err.Error()))
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Println(color.RedString("No choices found in response"))
		return "", errors.New("no choices found in response")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		log.Println(color.RedString("No message found in first choice"))
		return "", errors.New("no message found in first choice")
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Println(color.RedString("No content found in message"))
		return "", errors.New("no content found in message")
	}

	return strings.TrimSpace(content), nil

}

func WriteAltTextToFile(altText, outputPath string) error {

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Println(color.RedString("Failed to make directory to write output file: %s", err.Error()))
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(color.RedString("Failed to open output file to write results: %s", err.Error()))
		return fmt.Errorf("failed to open file %s: %w", outputPath, err)
	}
	defer file.Close()

	if _, err = file.WriteString(altText); err != nil {
		log.Println(color.RedString("Failed to write output to file: %s", err.Error()))
		return fmt.Errorf("failed to write alt text to %s: %w", outputPath, err)
	}

	return nil

}

func PublishAltTextResultsToS3(altText string) (string, error) {

	keyID := uuid.New().String()

	if os.Getenv("AWS_REGION") == "" {
		return "", errors.New("AWS_REGION environment variable has not been set. Will not attempt upload to S3.")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		return "", errors.New("AWS_ACCESS_KEY_ID environment variable has not been set. Cannot attempt upload to S3.")
	}

	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		return "", errors.New("AWS_SECRET_ACCESS_KEY environment variable has not been set. Cannot attempt upload to S3.")
	}

	if os.Getenv("S3_BUCKET") == "" {
		return "", errors.New("S3_BUCKET environment variable has not been set. Cannot attempt upload to S3.")
	}

	uploader, uploaderErr := s3uploader.NewS3Uploader(os.Getenv("S3_BUCKET"))
	if uploaderErr != nil {
		return "", uploaderErr
	}

	content := []byte(altText)
	contentType := "text/plain"

	if _, uploadErr := uploader.UploadFile(keyID, content, contentType); uploadErr != nil {
		return "", uploadErr
	}

	return keyID, nil

}
