package main

import (
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

var Jobs map[string]JobFactory

func init() {
	Jobs = map[string]JobFactory{
		"@every 1m": NewUpdateFeedsJob,
	}
}

type JobFactory func(conf *Config) cron.Job

type UpdateFeedsJob struct {
	conf *Config
}

func NewUpdateFeedsJob(conf *Config) cron.Job {
	return &UpdateFeedsJob{conf: conf}
}

func (job *UpdateFeedsJob) Run() {
	conf := job.conf
	for name, url := range conf.Feeds {
		if err := UpdateFeed(conf, name, url); err != nil {
			log.WithError(err).Errorf("error updating feed %s: %s", name, url)
		}
	}
}
