package netscaler

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

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

type Server struct {
	Name string
	IP   string
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

func (c *Client) GetServers() ([]Server, error) {
	servers := []Server{}
	req, err := c.request("GET", "config/server", nil)
	if err != nil {
		return servers, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return servers, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return servers, err
	}

	if res.StatusCode != 200 {
		return servers, errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
	}

	err = jp.ArrayEach(body, func(value []byte, dataType jp.ValueType, offset int, err error) {
		name, err := jp.GetString(value, "name")
		if err != nil {
			return
		}
		ip, err := jp.GetString(value, "ipaddress")
		if err != nil {
			return
		}
		servers = append(servers, Server{
			Name: name,
			IP:   ip,
		})
	}, "server")

	return servers, err
}

func (c *Client) AddServers(servers ...Server) error {
	list := []string{}
	for _, server := range servers {
		list = append(list,
			fmt.Sprintf(`{"name": "%s", "ipaddress": "%s"}`, server.Name, server.IP))
	}

	req, err := c.request("POST", "config/server",
		strings.NewReader(fmt.Sprintf(`{"server": [%s]}`, strings.Join(list, ","))))
	req.Header.Set("Content-Type", "application/vnd.com.citrix.netscaler.server_list+json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Unexpected HTTP status: %d", res.StatusCode))
	}

	return nil
}

func (c *Client) RemoveServers(names ...string) error {
	for _, name := range names {
		req, err := c.request("DELETE", fmt.Sprintf("config/server/%s", name), nil)
		if err != nil {
			return err
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Unexpected HTTP status: %d while deleting %s", res.StatusCode, name))
		}
	}

	return nil
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
