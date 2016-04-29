package netscaler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	jp "github.com/buger/jsonparser"
)

type Config struct {
	URL               string
	HTTPClient        *http.Client
	HTTPBasicAuthUser string
	HTTPBasicPassword string
}

type Client struct {
	config     *Config
	httpClient *http.Client
	version    string
}

func NewClient(config *Config) *Client {
	httpClient := http.DefaultClient
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
	}
}

func (c *Client) Version() (string, error) {
	if c.version != "" {
		return c.version, nil
	}

	req, err := c.request("GET", "config/nsversion", nil)
	if err != nil {
		return "", err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
	}

	c.version, err = jp.GetString(body, "nsversion", "version")
	return c.version, err
}

func (c *Client) create(resourceType string, resource interface{}) error {
	var buffer []byte
	buffer, err := json.Marshal(&resource)
	if err != nil {
		return err
	}

	apiQuery := fmt.Sprintf("config/%s", resourceType)
	req, err := c.request("POST", apiQuery, bytes.NewReader(buffer))
	if err != nil {
		return err
	}

	contentType := fmt.Sprintf("application/vnd.com.citrix.netscaler.%s+json", resourceType)
	req.Header.Set("Content-Type", contentType)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
	}

	return nil
}

func (c *Client) delete(resourceType, resourceName string) error {
	apiQuery := fmt.Sprintf("config/%s/%s", resourceType, resourceName)

	req, err := c.request("DELETE", apiQuery, nil)
	if err != nil {
		return err
	}

	contentType := fmt.Sprintf("application/vnd.com.citrix.netscaler.%s+json", resourceType)
	req.Header.Set("Content-Type", contentType)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
	}

	return nil
}

func (c *Client) query(resourceType, filter string, result interface{}) error {
	uriParams := ""
	if filter != "" {
		uriParams = "?filter=" + url.QueryEscape(filter)
	}

	apiQuery := fmt.Sprintf("config/%s%s", resourceType, uriParams)
	req, err := c.request("GET", apiQuery, nil)
	if err != nil {
		return err
	}

	contentType := fmt.Sprintf("application/vnd.com.citrix.netscaler.%s+json", resourceType)
	req.Header.Set("Content-Type", contentType)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		if res.StatusCode == 400 {
			return nil
		}

		data, _, _, err := jp.Get(body, resourceType)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, result)
	}

	return errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
}

func (c *Client) request(method, path string, body io.Reader) (*http.Request, error) {
	uri := fmt.Sprintf("%v/nitro/v1/%s", c.config.URL, path)
	req, err := http.NewRequest("GET", uri, body)
	if err != nil {
		return nil, err
	}

	if c.config.HTTPBasicAuthUser != "" {
		req.SetBasicAuth(c.config.HTTPBasicAuthUser, c.config.HTTPBasicPassword)
	}

	return req, nil
}
