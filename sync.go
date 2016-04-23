package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
	"github.com/gettyimages/marathon-netscaler/netscaler"
)

type SyncHandler struct {
	reader *SyncReader
	writer *SyncWriter
}

type SyncReader struct {
	mc marathon.Marathon
}

type SyncWriter struct {
	nc *netscaler.Client
}

func newSyncHandler(mc marathon.Marathon, nc *netscaler.Client) *SyncHandler {
	return &SyncHandler{
		reader: &SyncReader{mc},
		writer: &SyncWriter{nc},
	}
}

func (s *SyncHandler) Do() {

	apps, err := s.reader.Apps()
	if err != nil {
		log.Errorf("Problem getting apps to sync: %v", err)
		return
	}

	log.Printf("Found %d apps to sync...", len(apps))

	err = s.writer.Apps(apps)
	if err != nil {
		log.Errorf("Problem syncing latest apps: %v", err)
		return
	}

	log.Printf("Done syncing %d apps.", len(apps))
}

func (s *SyncHandler) Close() {
}

func (r *SyncReader) Apps() ([]*App, error) {
	var results []*App
	apps, _ := r.mc.Applications(nil)
	tasks, _ := r.mc.AllTasks(&marathon.AllTasksOpts{"running"})

	for _, app := range apps.Apps {
		if app.Labels == nil {
			continue
		}

		labels := *app.Labels
		serverGroup, ok := labels["netscaler.service_group"]
		if !ok {
			continue
		}

		results = append(results, &App{
			ID:           app.ID,
			ServiceGroup: serverGroup,
			Addrs:        []string{},
		})
	}

	for _, task := range tasks.Tasks {
		for _, res := range results {
			if res.ID != task.AppID {
				continue
			}
			if len(task.Ports) == 0 {
				continue
			}

			res.Addrs = append(res.Addrs,
				fmt.Sprintf("%s:%d", task.Host, task.Ports[0]))
		}
	}

	return results, nil
}

func (w *SyncWriter) Apps(apps []*App) error {
	return nil
}
