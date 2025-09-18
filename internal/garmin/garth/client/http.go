package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sstent/fitness-tui/internal/garmin/garth/errors"
)

func (c *Client) ConnectAPI(path, method string, data interface{}) (interface{}, error) {
	url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, path)

	var body io.Reader
	if data != nil && (method == "POST" || method == "PUT") {
		jsonData, err := json.Marshal(data)
		if err != nil {
			// Changed to IOError since this is a serialization error
			return nil, &errors.ValidationError{GarthError: errors.GarthError{
				Message: "Failed to marshal request data", Cause: err}}
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "Failed to create request", Cause: err}}}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "GCM-iOS-5.7.2.1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "API request failed", Cause: err}}}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			StatusCode: resp.StatusCode,
			Response:   string(bodyBytes),
			GarthError: errors.GarthError{Message: fmt.Sprintf("API error: %s", resp.Status)}}}
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Failed to parse response", Cause: err}}
	}

	return result, nil
}

func (c *Client) Download(path string) ([]byte, error) {
	url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Wrap with APIError for consistency
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "Failed to create download request", Cause: err}}}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "GCM-iOS-5.7.2.1")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "Download request failed", Cause: err}}}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			StatusCode: resp.StatusCode,
			Response:   string(bodyBytes),
			GarthError: errors.GarthError{Message: fmt.Sprintf("Download failed: %s", resp.Status)}}}
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Upload(filePath, uploadPath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Failed to open file", Cause: err}}
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		// Wrap with IOError
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Failed to create form file", Cause: err}}
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Failed to copy file data", Cause: err}}
	}
	writer.Close()

	url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, uploadPath)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "Failed to create upload request", Cause: err}}}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			GarthError: errors.GarthError{Message: "Upload request failed", Cause: err}}}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
			StatusCode: resp.StatusCode,
			Response:   string(bodyBytes),
			GarthError: errors.GarthError{Message: fmt.Sprintf("Upload failed: %s", resp.Status)}}}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Failed to parse upload response", Cause: err}}
	}

	return result, nil
}
