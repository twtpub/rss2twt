package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/andyleap/microformats"
	"github.com/gosimple/slug"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const (
	avatarResolution = 60 // 60x60 px
	rssTwtxtTemplate = "%s\t%s ⌘ [Read more...](%s)\n"
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

func TestFeed(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

func FindFeed(uri string) (*gofeed.Feed, string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, "", err
	}

	res, err := http.Get(u.String())
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	p := microformats.New()
	data := p.Parse(res.Body, u)

	altMap := make(map[string]string)
	for _, alt := range data.Alternates {
		altMap[alt.Type] = alt.URL
	}

	feedURL := altMap["application/atom+xml"]

	if feedURL == "" {
		for _, alt := range data.Alternates {
			switch alt.Type {
			case "application/atom+xml", "application/rss+xml":
				feedURL = alt.URL
				break
			}
		}
	}

	if feedURL == "" {
		return nil, "", ErrNoSuitableFeedsFound
	}

	feed, err := TestFeed(feedURL)
	if err != nil {
		return nil, "", err
	}

	return feed, feedURL, nil
}

// ValidateFeed ...
func ValidateFeed(conf *Config, url string) (Feed, error) {
	feed, err := TestFeed(url)
	if err != nil {
		log.WithError(err).Warnf("invalid feed %s", url)
	}

	if feed == nil {
		feed, url, err = FindFeed(url)
		if err != nil {
			log.WithError(err).Errorf("no feed found on %s", url)
			return Feed{}, err
		}
	}

	name := slug.Make(feed.Title)

	if feed.Image != nil && feed.Image.URL != "" {
		opts := &ImageOptions{
			Resize:  true,
			ResizeW: avatarResolution,
			ResizeH: avatarResolution,
		}

		filename := fmt.Sprintf("%s.png", name)

		if err := DownloadImage(conf, feed.Image.URL, filename, opts); err != nil {
			log.WithError(err).Warnf("error downloading feed image from %s", feed.Image.URL)
		}
	}

	return Feed{Name: name, URL: url}, nil
}

func UpdateFeed(conf *Config, name, url string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return err
	}

	avatarFile := filepath.Join(conf.Root, fmt.Sprintf("%s.png", name))
	if feed.Image != nil && feed.Image.URL != "" && !Exists(avatarFile) {
		opts := &ImageOptions{
			Resize:  true,
			ResizeW: avatarResolution,
			ResizeH: avatarResolution,
		}

		filename := fmt.Sprintf("%s.png", name)

		if err := DownloadImage(conf, feed.Image.URL, filename, opts); err != nil {
			log.WithError(err).Warnf("error downloading feed image from %s", feed.Image.URL)
		}
	}

	var lastModified = time.Time{}

	fn := filepath.Join(conf.Root, fmt.Sprintf("%s.txt", name))

	stat, err := os.Stat(fn)
	if err == nil {
		lastModified = stat.ModTime()
	}

	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	old, new := 0, 0
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		if item.PublishedParsed.After(lastModified) {
			new++
			text := fmt.Sprintf(
				rssTwtxtTemplate,
				item.PublishedParsed.Format(time.RFC3339),
				item.Title,
				item.Link,
			)
			_, err := f.WriteString(text)
			if err != nil {
				return err
			}
		} else {
			old++
		}
	}

	if (old + new) == 0 {
		log.WithField("name", name).WithField("url", url).Warn("empty or bad feed")
	}

	return nil
}
