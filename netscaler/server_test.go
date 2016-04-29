package netscaler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	jp "github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_netscaler_find_servers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/nitro/v1/config/server",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"server": [{
					"name": "server1",
					"ipaddress": "1.1.1.1"
				}, {
					"name": "server2",
					"ipaddress": "2.2.2.2"
				}]
			}`)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := &Config{
		URL: server.URL,
	}

	client := NewClient(config)
	servers, err := client.GetServers("")

	require.Nil(err)
	require.Equal(2, len(servers))

	assert.Equal("server1", servers[0].Name)
	assert.Equal("1.1.1.1", servers[0].IP)

	assert.Equal("server2", servers[1].Name)
	assert.Equal("2.2.2.2", servers[1].IP)
}

func Test_netscaler_add_servers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	servers := []string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/nitro/v1/config/server",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			name, _ := jp.GetString(body, "name")
			servers = append(servers, name)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := &Config{
		URL: server.URL,
	}
	client := NewClient(config)
	err := client.AddServers([]Server{
		NewServer("server1", "1.1.1.1"),
		NewServer("server2", "2.2.2.2"),
	})

	require.Nil(err)
	require.Equal(2, len(servers))
	assert.Equal("server1", servers[0])
	assert.Equal("server2", servers[1])
}

func Test_netscaler_remove_servers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	deleted := []bool{}
	mux := http.NewServeMux()
	mux.HandleFunc("/nitro/v1/config/server/server1",
		func(w http.ResponseWriter, r *http.Request) {
			deleted = append(deleted, true)
		})
	mux.HandleFunc("/nitro/v1/config/server/server2",
		func(w http.ResponseWriter, r *http.Request) {
			deleted = append(deleted, true)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := &Config{
		URL: server.URL,
	}
	client := NewClient(config)
	err := client.RemoveServers([]string{"server1", "server2"})

	require.Nil(err)
	assert.Equal(2, len(deleted))
}
