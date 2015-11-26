package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/jc-m/test-docker-plugin/routed/driver"
	"github.com/jc-m/test-docker-plugin/routed/server"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var (
		address string
		version string
	)

	flag.StringVar(&address, "socket", "/run/docker/plugins/routed.sock", "socket on which to listen")

	flag.Parse()

	log.Info("Test routed network plugin")
	version = "1"
	var d server.Driver
	d, err := driver.New(version)
	if err != nil {
		log.Fatalf("unable to create driver: %s", err)
	}
	var listener net.Listener

	// remove socket from last invocation
	if err := os.Remove(address); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	listener, err = net.Listen("unix", address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	endChan := make(chan error, 1)
	go func() {
		endChan <- server.Listen(listener, d)
	}()

	select {
	case sig := <-sigChan:
		log.Debugf("Caught signal %s; shutting down", sig)
	case err := <-endChan:
		if err != nil {
			log.Errorf("Error from listener: ", err)
			listener.Close()
			os.Exit(1)
		}
	}
}
