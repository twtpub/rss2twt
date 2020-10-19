package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// Blank import so we can handle image/*
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	"github.com/JesusIslam/tldr"
	"github.com/h2non/filetype"
	"github.com/k3a/html2text"
	shortuuid "github.com/lithammer/shortuuid/v3"
	"github.com/mmcdole/gofeed"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

var (
	validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\- ]*$`)

	ErrInvalidName  = errors.New("error: invalid feed name")
	ErrNameTooLong  = errors.New("error: name is too long")
	ErrInvalidImage = errors.New("error: invalid image")
)

func GetContent(item *gofeed.Item) string {
	var text string

	if item.Description != "" {
		text = item.Description
	} else if item.Content != "" {
		text = item.Content
	}

	text = html2text.HTML2Text(text)

	bag := tldr.New()
	result, err := bag.Summarize(text, 3)
	if err != nil {
		log.WithError(err).Error("error summarizing content")
		return ""
	}

	return CleanTwt(strings.Join(result, "\n"))
}

// CleanTwt cleans a twt's text, replacing new lines with spaces and
// stripping surrounding spaces.
func CleanTwt(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", "\u2028")
	return text
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsImage(fn string) bool {
	f, err := os.Open(fn)
	if err != nil {
		log.WithError(err).Warnf("error opening file %s", fn)
		return false
	}
	defer f.Close()

	head := make([]byte, 261)
	if _, err := f.Read(head); err != nil {
		log.WithError(err).Warnf("error reading from file %s", fn)
		return false
	}

	if filetype.IsImage(head) {
		return true
	}

	return false
}

type ImageOptions struct {
	Resize  bool
	ResizeW int
	ResizeH int
}

func DownloadImage(conf *Config, url string, resource, name string, opts *ImageOptions) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		log.WithError(err).Errorf("error downloading image from %s", url)
		return "", err
	}
	defer res.Body.Close()

	tf, err := ioutil.TempFile("", "rss2twtxt-*")
	if err != nil {
		log.WithError(err).Error("error creating temporary file")
		return "", err
	}
	defer tf.Close()

	if _, err := io.Copy(tf, res.Body); err != nil {
		log.WithError(err).Error("error writng temporary file")
		return "", err
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return "", err
	}

	if !IsImage(tf.Name()) {
		return "", ErrInvalidImage
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return "", err
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return "", err
	}

	img, _, err := image.Decode(tf)
	if err != nil {
		log.WithError(err).Error("jpeg.Decode failed")
		return "", err
	}

	newImg := img

	if opts != nil {
		if opts.Resize && (opts.ResizeW+opts.ResizeH > 0) && (opts.ResizeH > 0 || img.Bounds().Size().X > opts.ResizeW) {
			newImg = resize.Resize(uint(opts.ResizeW), uint(opts.ResizeH), img, resize.Lanczos3)
		}
	}

	p := filepath.Join(conf.Root, resource)
	if err := os.MkdirAll(p, 0755); err != nil {
		log.WithError(err).Error("error creating avatars directory")
		return "", err
	}

	var fn string

	if name == "" {
		uuid := shortuuid.New()
		fn = filepath.Join(p, fmt.Sprintf("%s.png", uuid))
	} else {
		fn = fmt.Sprintf("%s.png", filepath.Join(p, name))
	}

	of, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.WithError(err).Error("error opening output file")
		return "", err
	}
	defer of.Close()

	if err := png.Encode(of, newImg); err != nil {
		log.WithError(err).Error("error encoding image")
		return "", err
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		strings.TrimSuffix(conf.BaseURL, "/"),
		resource, strings.TrimSuffix(filepath.Base(fn), filepath.Ext(fn)),
	), nil
}
