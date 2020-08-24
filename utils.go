package main

import (
	"errors"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	// Blank import so we can handle image/*
	_ "image/gif"
	_ "image/jpeg"

	"github.com/h2non/filetype"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

var (
	validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\- ]*$`)

	ErrInvalidName  = errors.New("error: invalid feed name")
	ErrNameTooLong  = errors.New("error: name is too long")
	ErrInvalidImage = errors.New("error: invalid image")
)

func Exists(name string) bool {
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

func DownloadImage(conf *Config, url string, filename string, opts *ImageOptions) error {
	res, err := http.Get(url)
	if err != nil {
		log.WithError(err).Errorf("error downloading image from %s", url)
		return err
	}
	defer res.Body.Close()

	tf, err := ioutil.TempFile("", "rss2twtxt-*")
	if err != nil {
		log.WithError(err).Error("error creating temporary file")
		return err
	}
	defer tf.Close()

	if _, err := io.Copy(tf, res.Body); err != nil {
		log.WithError(err).Error("error writng temporary file")
		return err
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return err
	}

	if !IsImage(tf.Name()) {
		return ErrInvalidImage
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return err
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return err
	}

	img, _, err := image.Decode(tf)
	if err != nil {
		log.WithError(err).Error("jpeg.Decode failed")
		return err
	}

	newImg := img

	if opts != nil {
		if opts.Resize && (opts.ResizeW+opts.ResizeH > 0) && (opts.ResizeH > 0 || img.Bounds().Size().X > opts.ResizeW) {
			newImg = resize.Resize(uint(opts.ResizeW), uint(opts.ResizeH), img, resize.Lanczos3)
		}
	}

	fn := filepath.Join(conf.Root, filename)

	of, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.WithError(err).Error("error opening output file")
		return err
	}
	defer of.Close()

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	if err := png.Encode(of, newImg); err != nil {
		log.WithError(err).Error("error reencoding image")
		return err
	}

	return nil
}
