package golibxml

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Client represents a client for the API.
type Client struct {
	apiKey  string
	token   string
	baseURL string
}

// Node represents a node in an XML file.
type Node struct {
	XMLName string `json:"xmlname"`
	Value   string `json:"value"`
	Nodes   []Node `json:"nodes"`
}

// APIResponse represents a response from the API.
type APIResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

// NewClient creates a new client.
func NewClient(apiKey, baseURL string) *Client {
	return &Client{apiKey: apiKey, baseURL: baseURL}
}

// request makes a request to the API.
func (c *Client) request(method, endpoint string, params map[string]string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add query parameters
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Add body if present
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New(string(respBody))
	}

	return respBody, nil
}

// Authorize authorizes the client with the API.
func (c *Client) Authorize() error {
	resp, err := c.request("GET", "/authorize", nil, nil)
	if err != nil {
		return err
	}

	var result map[string]string
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return err
	}

	if token, ok := result["token"]; ok {
		c.token = token
		return nil
	}

	return errors.New("authorization failed")
}

// CopyDevice copies a device. If a filename is provided only that file is copied, otherwise all files are copied.
func (c *Client) CopyDevice(deviceID, newDeviceID, filename string, overwrite bool) (string, error) {
	params := map[string]string{
		"deviceid":     deviceID,
		"new_deviceid": newDeviceID,
		"filename":     filename,
		"overwrite":    fmt.Sprintf("%t", overwrite),
	}

	resp, err := c.request("POST", "/copyDevice", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// CreateFile creates a new XML file.
func (c *Client) CreateFile(deviceID, filename, rootName string) (string, error) {
	params := map[string]string{
		"deviceid": deviceID,
		"filename": filename,
		"rootname": rootName,
	}

	resp, err := c.request("POST", "/createFile", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// CreateNode creates a new node in the XML file.
func (c *Client) CreateNode(deviceID, filename, parentPath, tag, value string) (string, error) {
	params := map[string]string{
		"deviceid":    deviceID,
		"filename":    filename,
		"parent_path": parentPath,
		"tag":         tag,
		"value":       value,
	}

	resp, err := c.request("POST", "/create", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// DeleteNode deletes a node from the XML file.
func (c *Client) DeleteNode(deviceID, filename, path string) (string, error) {
	params := map[string]string{
		"deviceid": deviceID,
		"filename": filename,
		"path":     path,
	}

	resp, err := c.request("DELETE", "/delete", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// DeleteFile deletes an XML file.
func (c *Client) DeleteFile(deviceID, filename string) (string, error) {
	params := map[string]string{
		"deviceid": deviceID,
		"filename": filename,
	}

	resp, err := c.request("DELETE", "/deleteFile", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// ListFiles lists the XML files for a device.
func (c *Client) ListFiles(deviceID string) (string, error) {
	params := map[string]string{
		"deviceid": deviceID,
	}

	resp, err := c.request("GET", "/listFile", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}

// ReadNode reads a node from the XML file.
func (c *Client) ReadNode(deviceID, filename, path string) (*Node, error) {
	params := map[string]string{
		"deviceid": deviceID,
		"filename": filename,
		"path":     path,
	}

	resp, err := c.request("GET", "/read", params, nil)
	if err != nil {
		return nil, err
	}

	var node Node
	err = json.Unmarshal(resp, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

// UpdateNode updates a node in the XML file.
func (c *Client) UpdateNode(deviceID, filename, path, value string) (string, error) {
	params := map[string]string{
		"deviceid": deviceID,
		"filename": filename,
		"path":     path,
		"value":    value,
	}

	resp, err := c.request("PUT", "/update", params, nil)
	if err != nil {
		return "", err
	}

	var result APIResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(result.Error)
	}

	return result.Data, nil
}
