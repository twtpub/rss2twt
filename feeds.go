package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/andyleap/microformats"
	"github.com/gosimple/slug"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNoSuitableFeedsFound = errors.New("error: no suitable RSS or Atom feeds found")
)

// Feed ...
type Feed struct {
	Name string
	URL  string

	LastModified string
}

// ValidateFeed ...
func ValidateFeed(uri string) (Feed, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return Feed{}, err
	}

	res, err := http.Get(u.String())
	if err != nil {
		return Feed{}, err
	}
	defer res.Body.Close()

	p := microformats.New()
	data := p.Parse(res.Body, u)

	var feedURL string
	for _, alt := range data.Alternates {
		switch alt.Type {
		case "application/atom+xml", "application/rss+xml":
			feedURL = alt.URL
			break
		}
	}

	if feedURL == "" {
		return Feed{}, ErrNoSuitableFeedsFound
	}

	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return Feed{}, err
	}

	return Feed{
		Name: slug.Make(feed.Title),
		URL:  feedURL,
	}, nil
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
		if item.PublishedParsed == nil {
			log.WithField("item", item).Warn("item has no published date")
			continue
		}

		if item.PublishedParsed.After(lastModified) {
			text := fmt.Sprintf("%s\t%s ⌘ %s\n", item.PublishedParsed.Format(time.RFC3339), item.Title, item.Link)
			_, err := f.WriteString(text)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
