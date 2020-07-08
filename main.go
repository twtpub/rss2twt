package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	version bool
	debug   bool

	server bool
	bind   string
	config string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVarP(&version, "version", "v", false, "display version information")
	flag.BoolVarP(&debug, "debug", "d", false, "enable debug logging")

	flag.BoolVarP(&server, "server", "s", false, "enable server mode")
	flag.StringVarP(&bind, "bind", "b", "0.0.0.0:8000", "interface and port to bind to in server mode")
	flag.StringVarP(&config, "config", "c", "config.yaml", "configuration file for server mode")
}

func main() {
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if version {
		fmt.Printf("rss2twtxt %s\n", FullVersion())
		os.Exit(0)
	}

	if server {
		app, err := NewApp(bind, config)
		if err != nil {
			log.WithError(err).Fatal("error creating app for server mode")
		}
		if err := app.Run(); err != nil {
			log.WithError(err).Fatal("error running app")
		}
		os.Exit(0)
	}

	url := flag.Arg(0)
	name := flag.Arg(1)

	filename := fmt.Sprintf("%s.txt", name)

	if err := UpdateFeed(filename, url); err != nil {
		log.WithError(err).Fatal("error updating feed")
	}
}
