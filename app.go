package main

import (
	"net/http"

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

	router.HandleFunc("/", app.IndexHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/feeds", app.FeedsHandler).Methods(http.MethodGet)
	router.HandleFunc("/{name}/twtxt.txt", app.FeedHandler).Methods(http.MethodGet)

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
