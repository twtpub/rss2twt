package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type App struct {
	bind   string
	conf   *Config
	cron   *cron.Cron
	router *mux.Router
}

func NewApp(bind, config string) (*App, error) {
	conf, err := LoadConfig(config)
	if err != nil {
		return nil, err
	}

	cron := cron.New()

	return &App{
		bind: bind,
		conf: conf,
		cron: cron,
	}, nil
}

func (app *App) initRoutes() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", app.IndexHandler).Methods(http.MethodGet, http.MethodHead, http.MethodPost)
	router.HandleFunc("/feeds", app.FeedsHandler).Methods(http.MethodGet, http.MethodHead)
	router.HandleFunc("/we-are-feeds.txt", app.WeAreFeedsHandler).Methods(http.MethodGet, http.MethodHead)
	router.HandleFunc("/{name}/twtxt.txt", app.FeedHandler).Methods(http.MethodGet, http.MethodHead)
	router.HandleFunc("/{name}/avatar.png", app.AvatarHandler).Methods(http.MethodGet, http.MethodHead)

	return router
}

func (app *App) setupCronJobs() error {
	for spec, factory := range Jobs {
		job := factory(app.conf)
		if err := app.cron.AddJob(spec, job); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) GetFeeds() (feeds []Feed) {
	for name := range app.conf.Feeds {
		filename := filepath.Join(app.conf.Root, fmt.Sprintf("%s.txt", name))

		stat, err := os.Stat(filename)
		if err != nil {
			log.WithError(err).Warnf("error getting feed stats for %s", name)
			continue
		}
		lastModified := humanize.Time(stat.ModTime())

		url := fmt.Sprintf("%s/%s/twtxt.txt", app.conf.BaseURL, name)
		feeds = append(feeds, Feed{name, url, lastModified})
	}

	sort.Slice(feeds, func(i, j int) bool { return feeds[i].Name < feeds[j].Name })

	return
}

func (app *App) Run() error {
	router := app.initRoutes()

	if err := app.setupCronJobs(); err != nil {
		log.WithError(err).Error("error setting up background jobs")
		return err
	}

	app.cron.Start()
	log.Infof("started background jobs")

	log.Infof("rss2twtxt %s listening on http://%s", FullVersion(), app.bind)

	return http.ListenAndServe(app.bind, router)
}
