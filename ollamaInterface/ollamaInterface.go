package ollamaInterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Client is used to interact with the Ollama API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Ollama Client with the specified base URL.
// Example: NewClient("http://localhost:11434")
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// doPOST sends a POST request to path with the given body (encoded as JSON).
// It attempts to parse the response as one or more JSON objects.
// Returns a slice of map[string]interface{}, one entry per JSON object in the response.
func (c *Client) doPOST(path string, body map[string]interface{}) ([]map[string]interface{}, error) {
	// Marshal request body to JSON
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("json marshal error: %w", err)
		}
	}

	// Create and execute the request
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Read any error body
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-2xx response: %d, body: %s", resp.StatusCode, string(b))
	}

	// For streaming endpoints, multiple JSON objects might appear one after another.
	// We'll decode all of them until EOF.
	decoder := json.NewDecoder(resp.Body)
	var results []map[string]interface{}

	for {
		var m map[string]interface{}
		decodeErr := decoder.Decode(&m)
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			return results, fmt.Errorf("json decode error: %w", decodeErr)
		}
		results = append(results, m)
	}

	return results, nil
}

// doGET sends a GET request to the given path.
// It expects a single JSON object in response and returns map[string]interface{}.
func (c *Client) doGET(path string) (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("GET request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-2xx response: %d, body: %s", resp.StatusCode, string(b))
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Below are methods mapped to each Ollama API endpoint.
// ----------------------------------------------------------------------------

// GenerateCompletion calls POST /api/generate
// Returns one or more JSON objects (streaming responses).
func (c *Client) GenerateCompletion(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/generate", body)
}

// GenerateChatCompletion calls POST /api/chat
// Returns one or more JSON objects (streaming responses).
func (c *Client) GenerateChatCompletion(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/chat", body)
}

// CreateModel calls POST /api/create
// Returns streaming updates about model creation status.
func (c *Client) CreateModel(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/create", body)
}

// ListLocalModels calls GET /api/tags
// Returns a single JSON object with local model metadata.
func (c *Client) ListLocalModels() (map[string]interface{}, error) {
	return c.doGET("/api/tags")
}

// ShowModelInformation calls POST /api/show
// Typically returns a single JSON object with details about a model.
func (c *Client) ShowModelInformation(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/show", body)
}

// CopyModel calls POST /api/copy
// Copies a model from source to destination.
func (c *Client) CopyModel(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/copy", body)
}

// DeleteModel calls DELETE /api/delete
// Deletes the specified model.
func (c *Client) DeleteModel(body map[string]interface{}) ([]map[string]interface{}, error) {
	// Technically, the Ollama docs show "DELETE /api/delete" with a JSON body.
	// We'll do a POST with a method override or just do a NewRequest with method=DELETE.
	// We'll reuse doPOST-like logic but with method=DELETE, for simplicity:
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/delete", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DELETE request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-2xx response: %d, body: %s", resp.StatusCode, string(b))
	}

	// The response is typically a single JSON object, but let's decode as multiple
	// in case there's streaming (though the docs show a single object).
	decoder := json.NewDecoder(resp.Body)
	var results []map[string]interface{}
	for {
		var m map[string]interface{}
		decodeErr := decoder.Decode(&m)
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("json decode error: %w", decodeErr)
		}
		results = append(results, m)
	}

	return results, nil
}

// PullModel calls POST /api/pull
// Streams status updates about pulling the model.
func (c *Client) PullModel(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/pull", body)
}

// PushModel calls POST /api/push
// Streams status updates about pushing the model.
func (c *Client) PushModel(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/push", body)
}

// GenerateEmbeddings calls POST /api/embed
// Returns one or more JSON objects (though typically only one).
func (c *Client) GenerateEmbeddings(body map[string]interface{}) ([]map[string]interface{}, error) {
	return c.doPOST("/api/embed", body)
}

// ListRunningModels calls GET /api/ps
// Returns a single JSON object with loaded models.
func (c *Client) ListRunningModels() (map[string]interface{}, error) {
	return c.doGET("/api/ps")
}

// Version calls GET /api/version
// Returns a single JSON object with the version of Ollama.
func (c *Client) Version() (map[string]interface{}, error) {
	return c.doGET("/api/version")
}
