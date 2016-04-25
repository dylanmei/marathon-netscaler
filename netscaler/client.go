package netscaler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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

	uri := fmt.Sprintf("%v/nitro/v1/config/nsversion", c.config.URL)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}

	if c.config.HTTPBasicAuthUser != "" {
		req.SetBasicAuth(c.config.HTTPBasicAuthUser, c.config.HTTPBasicPassword)
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
