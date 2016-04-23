package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

type config struct {
	marathonUri  string
	netscalerUri string
	logLevel     string
}

func main() {
	config := config{}
	flag.StringVar(&config.marathonUri, "marathon.uri", "http://marathon.mesos:8080", "URI of Marathon API")
	flag.StringVar(&config.netscalerUri, "netscaler.uri", "http://netscaler.local", "URI of Netscaler API")
	flag.StringVar(&config.logLevel, "log.level", "info", "Logging level. Valid levels: [debug, info, warn, error, fatal].")
	flag.Parse()
	initLogger(config.logLevel)

	marathonUri, err := url.Parse(config.marathonUri)
	if err != nil {
		log.Fatal(err)
	}

	netscalerUri, err := url.Parse(config.netscalerUri)
	if err != nil {
		log.Fatal(err)
	}

	quitCh := make(chan os.Signal)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)

	marathonClient := marathonConnect(marathonUri, quitCh)
	netscalerClient := netscalerConnect(netscalerUri, quitCh)
	syncHandler := newSyncHandler(marathonClient, netscalerClient)
	syncHandler.Sync()

	eventPump, err := newEventPump(marathonClient)
	if err != nil {
		log.Fatal(err)
	}
	defer eventPump.Close()

runForever:
	for {
		select {
		case event := <-eventPump.Next:
			log.Printf("Got event %s: %s", color.GreenString(event.EventType), event.AppID)
			syncHandler.Sync()
		case <-quitCh:
			log.Print("Quitting...")
			syncHandler.Close()

			log.Print("Done.")
			break runForever
		}
	}
}
