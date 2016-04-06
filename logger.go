package main

import (
	log "github.com/Sirupsen/logrus"
)

func initLogger(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	log.SetLevel(l)
}
