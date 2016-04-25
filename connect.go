package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
	"github.com/gettyimages/marathon-netscaler/netscaler"
)

func netscalerConnect(netscalerUri *url.URL, quitCh <-chan os.Signal) *netscaler.Client {
	log.Printf("Connecting to NetScaler at URL: %v", netscalerUri)
	client, err := newNetscalerClient(netscalerUri)

	if err != nil {
		log.Debug(err.Error())
		for {
			select {
			case <-quitCh:
				log.Println("Quitting.")
				os.Exit(0)

			case <-time.After(10 * time.Second):
				log.Print("Retrying NetScaler...")
				client, err = newNetscalerClient(netscalerUri)

				if err == nil {
					break
				}

				log.Debug(err.Error())
			}
		}
	}

	return client
}

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

	return client
}

func newNetscalerClient(uri *url.URL) (*netscaler.Client, error) {
	config := &netscaler.Config{
		URL: uri.String(),
	}

	if uri.User != nil {
		config.HTTPBasicAuthUser = uri.User.Username()
		if pass, ok := uri.User.Password(); ok {
			config.HTTPBasicPassword = pass
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

	client := netscaler.NewClient(config)
	version, err := client.Version()

	if err == nil {
		log.Printf("Connected to NetScaler! Version=%s\n", version)
	}

	return client, err
}

func newMarathonClient(uri *url.URL) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = uri.String()
	config.EventsTransport = marathon.EventsTransportSSE

	if uri.User != nil {
		config.HTTPBasicAuthUser = uri.User.Username()
		if pass, ok := uri.User.Password(); ok {
			config.HTTPBasicPassword = pass
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

	client, err := marathon.NewClient(config)
	if err != nil {
		return nil, err
	}

	info, err := client.Info()
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to Marathon! Name=%s, Version=%s\n", info.Name, info.Version)
	return client, nil
}
