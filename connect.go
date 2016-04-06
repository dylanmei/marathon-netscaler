package main

import (
	"crypto/tls"
	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func marathonConnect(marathonUri *url.URL, quitCh <-chan os.Signal) marathon.Marathon {
	log.Printf("Connecting to Marathon at URL: %v", marathonUri)
	client, err := newMarathonClient(marathonUri)
	if err != nil {
		log.Debug(err.Error())
		for {
			select {
			case <-quitCh:
				log.Println("Quitting.")
				os.Exit(0)
			case <-time.After(10 * time.Second):
				log.Print("Retrying Marathon...")
				client, err = newMarathonClient(marathonUri)
				if err == nil {
					break
				}
				log.Debug(err.Error())
			}
		}
	}

	info, _ := client.Info()
	log.Printf("Connected to Marathon! Name=%s, Version=%s\n", info.Name, info.Version)

	return client
}

func newMarathonClient(uri *url.URL) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = uri.String()
	config.EventsTransport = marathon.EventsTransportSSE

	if uri.User != nil {
		if passwd, ok := uri.User.Password(); ok {
			config.HTTPBasicPassword = passwd
			config.HTTPBasicAuthUser = uri.User.Username()
		}
	}
	config.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return marathon.NewClient(config)
}
