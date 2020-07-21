package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type Feed struct {
	Name         string
	URL          string
	LastModified string
}

func ValidateFeed(url string) error {
	fp := gofeed.NewParser()
	_, err := fp.ParseURL(url)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFeed(filename, url string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return err
	}

	var lastModified = time.Time{}

	stat, err := os.Stat(filename)
	if err == nil {
		lastModified = stat.ModTime()
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, item := range feed.Items {
		if item.PublishedParsed.After(lastModified) {
			text := fmt.Sprintf("%s\t%s ⌘ %s\n", item.PublishedParsed.Format(time.RFC3339), item.Title, item.Link)
			n, err := f.WriteString(text)
			if err != nil {
				return err
			}
			log.Debugf("appended %d bytes to %s:\n%s", n, filename, text)
		}
	}

	return nil
}
