package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"text/template"

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

	http.Error(w, fmt.Sprintf("%s: %s", title, message), status)
	return nil
}

func (app *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx := struct {
			Title string
		}{
			Title: "rss2twtxt",
		}

		if err := render("index", indexTemplate, ctx, w); err != nil {
			log.WithError(err).Error("error rending index template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		name := r.FormValue("name")
		url := r.FormValue("url")

		if name == "" {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "No name supplied"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if url == "" {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "No url supplied"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if err := ValidateName(name); err != nil {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "Invalid feed name"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if err := ValidateFeed(url); err != nil {
			if err := renderMessage(w, http.StatusBadRequest, "Error", fmt.Sprintf("Invalid feed RSS/Atom feed: %s", url)); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		name = NormalizeName(name)

		if _, ok := app.conf.Feeds[name]; ok {
			if err := renderMessage(w, http.StatusConflict, "Error", "Feed alreadyd exists"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		app.conf.Feeds[name] = url
		if err := app.conf.Save(); err != nil {
			msg := fmt.Sprintf("Could not save feed: %s", err)
			if err := renderMessage(w, http.StatusInternalServerError, "Error", msg); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		msg := fmt.Sprintf("Feed successfully added %s: %s", name, url)
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
		Title: "Feeds",
		Feeds: app.GetFeeds(),
	}

	if err := render("feeds", feedsTemplate, ctx, w); err != nil {
		log.WithError(err).Error("error rendering feeds template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
