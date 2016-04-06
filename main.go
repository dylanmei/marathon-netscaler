package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	marathonUri string
	logLevel    string
}

func main() {
	config := config{}
	flag.StringVar(&config.marathonUri, "marathon.uri", "http://marathon.mesos:8080", "URI of Marathon")
	flag.StringVar(&config.logLevel, "log.level", "info", "Logging level. Valid levels: [debug, info, warn, error, fatal].")
	flag.Parse()
	initLogger(config.logLevel)

	marathonUri, err := url.Parse(config.marathonUri)
	if err != nil {
		log.Fatal(err)
	}

	quitCh := make(chan os.Signal)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)

	marathonClient := marathonConnect(marathonUri, quitCh)

	// TODO: reconcile existing apps

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
		case <-quitCh:
			log.Print("Quitting.")
			break runForever
		}
	}
}
