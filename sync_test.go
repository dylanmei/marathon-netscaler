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
					"ports": [123],
					"labels": {
						"netscaler.service_group": "foo"
					}
				}, {
					"id": "/service-without-label",
					"ports": [0]
				}, {
					"id": "/service-without-port",
					"labels": {
						"netscaler.service_group": "bar"
					}
				}]
			}`)
		})
	mux.HandleFunc("/v2/tasks",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"tasks": [{
					"id": "task-1",
					"appId": "/service-with-one-port",
					"slaveId": "agent-1",
					"host": "host-1",
					"ports": [123]
				}, {
					"id": "task-2",
					"appId": "/service-without-label",
					"slaveId": "agent-1",
					"host": "host-1",
					"ports": [456]
				}, {
					"id": "task-3",
					"appId": "/service-without-port",
					"slaveId": "agent-1",
					"host": "host-1",
					"ports": []
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

	assert.Equal(1, len(app.Agents))
	agent := app.Agents[0]
	assert.Equal("agent-1", agent.ID)
	assert.Equal("host-1", agent.Host)
}
