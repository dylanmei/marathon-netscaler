package netscaler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_netscaler_client_get_version(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/nitro/v1/config/nsversion",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"errorcode": 0,
				"nsversion": {
					"version": "foo"
				}
			}`)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := &Config{
		URL: server.URL,
	}
	client := NewClient(config)
	version, err := client.Version()

	require.Nil(err)
	assert.Equal("foo", version)
}
