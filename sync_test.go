package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sync_reader(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/apps",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"apps": [{
					"id": "/service-with-one-port",
					"labels": {
						"netscaler.service_group": "foo"
					}
				}, {
					"id": "/service-without-label"
				}]
			}`)
		})
	mux.HandleFunc("/v2/tasks",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"tasks": [{
					"id": "task-1",
					"appId": "/service-with-one-port",
					"host": "host-1",
					"ports": [4567]
				}, {
					"id": "task-2",
					"appId": "/service-without-label",
					"host": "host-1",
					"ports": [5678]
				}]
			}`)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := marathon.NewDefaultConfig()
	config.URL = server.URL
	client, _ := marathon.NewClient(config)

	reader := &SyncReader{client}
	apps, err := reader.Apps()

	require.NoError(err)
	assert.Equal(1, len(apps))

	app := apps[0]
	assert.Equal("/service-with-one-port", app.ID)
	assert.Equal("foo", app.ServiceGroup)

	assert.Equal(1, len(app.Addrs))
	assert.Equal("host-1:4567", app.Addrs[0])
}
