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
	avatarResolution = 60  // 60x60 px
	mediaResolution  = 640 // 640x480
	mediaDir         = "media"
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

	if feed.Image.URL != "" {
		opts := &ImageOptions{
			Resize:  true,
			ResizeW: avatarResolution,
			ResizeH: avatarResolution,
		}

		if _, err := DownloadImage(conf, feed.Image.URL, "", name, opts); err != nil {
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
	if feed.Image != nil && feed.Image.URL != "" && !FileExists(avatarFile) {
		opts := &ImageOptions{
			Resize:  true,
			ResizeW: avatarResolution,
			ResizeH: avatarResolution,
		}

		if _, err := DownloadImage(conf, feed.Image.URL, "", name, opts); err != nil {
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

		var mediaURI string

		if item.Image != nil && item.Image.URL != "" {
			opts := &ImageOptions{Resize: true, ResizeW: mediaResolution, ResizeH: 0}

			uri, err := DownloadImage(conf, item.Image.URL, mediaDir, "", opts)
			if err != nil {
				log.WithError(err).Warnf("error downloading item image from %s", item.Image.URL)
			} else {
				mediaURI = uri
			}
		}

		if item.PublishedParsed.After(lastModified) {
			new++

			timestamp := item.PublishedParsed.Format(time.RFC3339)

			content := fmt.Sprintf("**%s**", item.Title)

			if item.Description != "" || item.Content != "" {
				content += fmt.Sprintf("\u2028\u2028> %s", GetContent(item))
			}

			if mediaURI != "" {
				content += fmt.Sprintf("\u2028\u2028📷 ![%s](%s)", item.Image.Title, mediaURI)
			}

			content += fmt.Sprintf("\u2028\u2028👓 [Read more...](%s)", item.Link)

			line := fmt.Sprintf("%s\t%s\n", timestamp, content)

			_, err := f.WriteString(line)
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
