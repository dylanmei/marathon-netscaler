package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
	"github.com/gettyimages/marathon-netscaler/netscaler"
)

type SyncHandler struct {
	reader *SyncReader
	writer *SyncWriter
	sync   func()
}

type SyncReader struct {
	mc marathon.Marathon
}

type SyncWriter struct {
	nc *netscaler.Client
}

func newSyncHandler(mc marathon.Marathon, nc *netscaler.Client) *SyncHandler {
	reader := &SyncReader{mc}
	writer := &SyncWriter{nc}
	return &SyncHandler{
		reader: reader,
		writer: writer,
		sync: debounce(500*time.Millisecond, func() {
			apps, err := reader.Apps()
			if err != nil {
				log.Errorf("Problem finding apps to sync: %v", err)
				return
			}

			log.Printf("Found %d apps to sync...", len(apps))

			err = writer.Apps(apps)
			if err != nil {
				log.Errorf("Problem syncing apps: %v", err)
				return
			}

			log.Printf("Done syncing %d apps.", len(apps))
		}),
	}
}

func (s *SyncHandler) Do() {
	s.sync()
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

func debounce(interval time.Duration, f func()) func() {
	input := make(chan struct{})
	timer := time.NewTimer(interval)

	go func() {
		var ok bool

		// Do not start waiting for interval until called at least once
		_, ok = <-input

		// Channel closed; exit
		if !ok {
			return
		}

		// We start waiting for an interval
		for {
			select {

			case <-timer.C:
				// Interval has passed and we have a signal, so send it
				f()

				// Wait for another signal before waiting for an interval
				_, ok = <-input
				if !ok {
					return
				}

				timer.Reset(interval)

			case _, ok = <-input:
				// Channel closed; exit
				if !ok {
					return
				}
			}
		}
	}()

	return func() {
		input <- struct{}{}
	}
}