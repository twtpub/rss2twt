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
		fmt.Fprintf(os.Stderr, "用法: %s [配置项]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVarP(&version, "version", "v", false, "显示版本信息")
	flag.BoolVarP(&debug, "debug", "d", false, "启用调试")

	flag.BoolVarP(&server, "server", "s", false, "Web 服务模式")
	flag.StringVarP(&bind, "bind", "b", "0.0.0.0:8001", "Web 服务模式绑定地址及端口")
	flag.StringVarP(&config, "config", "c", "config.yaml", "Web 服务模式下使用的配置文件")
}

func main() {
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if version {
		fmt.Printf("rss2twt%s\n", FullVersion())
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

	if err := UpdateFeed(&Config{Root: "."}, name, url); err != nil {
		log.WithError(err).Fatal("error updating feed")
	}
}
