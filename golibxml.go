package xmlapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	apiKey  string
	token   string
	baseURL string
}

type XMLName struct {
	Space string `json:"Space"`
	Local string `json:"Local"`
}

type Node struct {
	XMLName XMLName `json:"XMLName"`
	Value   string  `json:"Value"`
	Nodes   []Node  `json:"Nodes"`
}

type APIResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{apiKey: apiKey, baseURL: baseURL}
}

func (c *Client) request(method, endpoint string, params map[string]string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	if endpoint == "/authorize" {
		req.Header.Set("Authorization", c.apiKey)
	} else {
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
			return
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		log.Printf("Request to %s failed with status: %d, response: %s", url, resp.StatusCode, respBody)
		return nil, errors.New(string(respBody))
	}

	return respBody, nil
}

func (c *Client) Authorize() error {
	url := fmt.Sprintf("%s%s", c.baseURL, "/authorize")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing body:", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		log.Printf("Authorization request failed with status: %d, response: %s", resp.StatusCode, respBody)
		return errors.New(string(respBody))
	}

	var result map[string]string
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return err
	}

	if token, ok := result["token"]; ok {
		c.token = token
		return nil
	}

	return errors.New("authorization failed")
}

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
