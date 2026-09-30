package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// APIClient handles communication with the Armur Code Scanner API.
type APIClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// WithAPIKey configures the client to authenticate requests with the given key.
func (c *APIClient) WithAPIKey(key string) *APIClient {
	c.APIKey = key
	return c
}

// do executes an HTTP request, injecting Authorization header when an API key is set.
func (c *APIClient) do(req *http.Request) (*http.Response, error) {
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	return c.HTTPClient.Do(req)
}

func (c *APIClient) postJSON(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *APIClient) getJSON(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// ScanRepository initiates a scan of a remote Git repository.
func (c *APIClient) ScanRepository(repoURL, language string, isAdvanced bool) (string, error) {
	var endpoint string
	if isAdvanced {
		endpoint = "/api/v1/advanced-scan/repo"
	} else {
		endpoint = "/api/v1/scan/repo"
	}
	fullURL := strings.TrimRight(c.BaseURL, "/") + endpoint

	body, err := json.Marshal(map[string]string{
		"repository_url": repoURL,
		"language":       language,
	})
	if err != nil {
		return "", fmt.Errorf("error creating request body: %w", err)
	}

	resp, err := c.postJSON(fullURL, body)
	if err != nil {
		return "", fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding API response: %w", err)
	}

	taskID, ok := result["task_id"]
	if !ok {
		return "", fmt.Errorf("task_id not found in API response")
	}

	return taskID, nil
}

// ScanFile initiates a scan of a local file via the file upload endpoint.
func (c *APIClient) ScanFile(filePath string, isAdvanced bool) (string, error) {
	fullURL := strings.TrimRight(c.BaseURL, "/") + "/api/v1/scan/file"

	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", filePath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory: directory scans are not supported by the CLI yet, scan a single file or a repository URL", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", filePath, err)
	}
	defer file.Close()

	// The endpoint expects a multipart upload with the file in the "file" field.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("error creating request body: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("error creating request body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error creating request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, &body)
	if err != nil {
		return "", fmt.Errorf("error making API request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding API response: %w", err)
	}

	taskID, ok := result["task_id"]
	if !ok {
		return "", fmt.Errorf("task_id not found in API response")
	}

	return taskID, nil
}

// GetTaskStatus retrieves the status of a specific scan task.
func (c *APIClient) GetTaskStatus(taskID string) (string, map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/status/%s", strings.TrimRight(c.BaseURL, "/"), taskID)

	resp, err := c.getJSON(url)
	if err != nil {
		return "", nil, fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("error decoding API response: %w", err)
	}

	status, ok := result["status"].(string)
	if !ok {
		return "", nil, fmt.Errorf("status not found in API response")
	}

	var data map[string]interface{}
	if status == "success" {
		data, ok = result["data"].(map[string]interface{})
		if !ok {
			return status, nil, fmt.Errorf("data not found or not a map in API response")
		}
	}

	return status, data, nil
}

// ProgressUpdate represents a real-time progress event from the server.
type ProgressUpdate struct {
	TaskID     string         `json:"task_id"`
	Status     string         `json:"status"`
	TotalTools int            `json:"total_tools"`
	Completed  int            `json:"completed"`
	StartedAt  int64          `json:"started_at"`
	Tools      []ToolProgress `json:"tools"`
}

// ToolProgress represents the status of a single tool.
type ToolProgress struct {
	Tool      string  `json:"tool"`
	Status    string  `json:"status"`
	StartedAt int64   `json:"started_at,omitempty"`
	EndedAt   int64   `json:"ended_at,omitempty"`
	Findings  int     `json:"findings"`
	Error     string  `json:"error,omitempty"`
	Duration  float64 `json:"duration_secs,omitempty"`
}

// StreamProgress connects to the SSE progress endpoint and sends updates to a channel.
// The channel is closed when the scan completes or an error occurs.
func (c *APIClient) StreamProgress(taskID string, updates chan<- ProgressUpdate) {
	defer close(updates)

	url := fmt.Sprintf("%s/api/v1/progress/%s", strings.TrimRight(c.BaseURL, "/"), taskID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "text/event-stream")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	var partial string
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			partial += string(buf[:n])
			// Process complete SSE messages
			for {
				idx := strings.Index(partial, "\n\n")
				if idx < 0 {
					break
				}
				message := partial[:idx]
				partial = partial[idx+2:]

				// Parse SSE data field
				for _, line := range strings.Split(message, "\n") {
					if strings.HasPrefix(line, "data: ") {
						data := strings.TrimPrefix(line, "data: ")
						var update ProgressUpdate
						if jsonErr := json.Unmarshal([]byte(data), &update); jsonErr == nil {
							updates <- update
							if update.Status == "completed" || update.Status == "failed" {
								return
							}
						}
					}
				}
			}
		}
		if err != nil {
			return
		}
	}
}

// GetOwaspReport retrieves the OWASP report for a completed scan task.
func (c *APIClient) GetOwaspReport(taskID string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/reports/owasp/%s", strings.TrimRight(c.BaseURL, "/"), taskID)

	resp, err := c.getJSON(url)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var report interface{}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("error decoding API response: %w", err)
	}

	return report, nil
}

// GetSansReport retrieves the SANS report for a completed scan task.
func (c *APIClient) GetSansReport(taskID string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/reports/sans/%s", strings.TrimRight(c.BaseURL, "/"), taskID)

	resp, err := c.getJSON(url)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var report interface{}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("error decoding API response: %w", err)
	}

	return report, nil
}
