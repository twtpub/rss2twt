package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/aofei/cameron"
	"github.com/gorilla/mux"
	"github.com/rickb777/accept"
	log "github.com/sirupsen/logrus"
)

func render(name, tmpl string, ctx interface{}, w io.Writer) error {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(w, ctx)
}

func renderMessage(w http.ResponseWriter, status int, title, message string) error {
	ctx := struct {
		Title   string
		Message string
	}{
		Title:   title,
		Message: message,
	}

	if err := render("message", messageTemplate, ctx, w); err != nil {
		return err
	}

	return nil
}

func (app *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx := struct {
			Title string
		}{
			Title: "RSS/Atom to twtxt feed aggregator service",
		}

		if err := render("index", indexTemplate, ctx, w); err != nil {
			log.WithError(err).Error("error rending index template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		url := r.FormValue("url")

		if url == "" {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "No url supplied"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		feed, err := ValidateFeed(app.conf, url)
		if err != nil {
			if err := renderMessage(w, http.StatusBadRequest, "Error", fmt.Sprintf("Unable to find a valid RSS/Atom feed for: %s", url)); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if _, ok := app.conf.Feeds[feed.Name]; ok {
			if err := renderMessage(w, http.StatusConflict, "Error", "Feed alreadyd exists"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		app.conf.Feeds[feed.Name] = feed.URL
		if err := app.conf.Save(); err != nil {
			msg := fmt.Sprintf("Could not save feed: %s", err)
			if err := renderMessage(w, http.StatusInternalServerError, "Error", msg); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		msg := fmt.Sprintf("Feed successfully added %s: %s", feed.Name, feed.URL)
		if err := renderMessage(w, http.StatusCreated, "Success", msg); err != nil {
			log.WithError(err).Error("error rendering message template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) FeedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	filename := filepath.Join(app.conf.Root, fmt.Sprintf("%s.txt", name))

	http.ServeFile(w, r, filename)
}

func (app *App) AvatarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, no-cache, must-revalidate")

	vars := mux.Vars(r)

	name := vars["name"]
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fn := filepath.Join(app.conf.Root, fmt.Sprintf("%s.png", name))
	if fileInfo, err := os.Stat(fn); err == nil {
		etag := fmt.Sprintf("W/\"%s-%s\"", r.RequestURI, fileInfo.ModTime().Format(time.RFC3339))

		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Header().Set("Etag", etag)
		if r.Method == http.MethodHead {
			return
		}

		f, err := os.Open(fn)
		if err != nil {
			log.WithError(err).Error("error opening avatar file")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(w, f); err != nil {
			log.WithError(err).Error("error writing avatar response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		return
	}

	etag := fmt.Sprintf("W/\"%s\"", r.RequestURI)

	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Etag", etag)
	if r.Method == http.MethodHead {
		return
	}

	buf := bytes.Buffer{}
	img := cameron.Identicon([]byte(name), avatarResolution, 12)
	png.Encode(&buf, img)

	w.Write(buf.Bytes())
}

func (app *App) WeAreFeedsHandler(w http.ResponseWriter, r *http.Request) {
	for _, feed := range app.GetFeeds() {
		fmt.Fprintf(w, "%s %s\n", feed.Name, feed.URL)
	}
}

func (app *App) FeedsHandler(w http.ResponseWriter, r *http.Request) {
	if accept.PreferredContentTypeLike(r.Header, "text/plain") == "text/plain" {
		app.WeAreFeedsHandler(w, r)
		return
	}

	ctx := struct {
		Title string
		Feeds []Feed
	}{
		Title: "Available twtxt feeds",
		Feeds: app.GetFeeds(),
	}

	if err := render("feeds", feedsTemplate, ctx, w); err != nil {
		log.WithError(err).Error("error rendering feeds template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
