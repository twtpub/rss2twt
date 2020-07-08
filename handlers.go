package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
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
			fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		name := r.FormValue("name")
		url := r.FormValue("url")

		if name == "" {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "No name supplied"); err != nil {
				fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if url == "" {
			if err := renderMessage(w, http.StatusBadRequest, "Error", "No url supplied"); err != nil {
				fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// TODO: Validate Name/URL for validity/length

		if _, ok := app.conf.Feeds[name]; ok {
			if err := renderMessage(w, http.StatusConflict, "Error", "Feed alreadyd exists"); err != nil {
				fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		app.conf.Feeds[name] = url
		if err := app.conf.Save(); err != nil {
			msg := fmt.Sprintf("Could not save feed: %s", err)
			if err := renderMessage(w, http.StatusInternalServerError, "Error", msg); err != nil {
				fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		msg := fmt.Sprintf("Feed successfully added %s: %s", name, url)
		if err := renderMessage(w, http.StatusCreated, "Success", msg); err != nil {
			fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
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

func (app *App) FeedsHandler(w http.ResponseWriter, r *http.Request) {
	var feeds []Feed

	for name, _ := range app.conf.Feeds {
		filename := filepath.Join(app.conf.Root, fmt.Sprintf("%s.txt", name))

		stat, err := os.Stat(filename)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		lastModified := humanize.Time(stat.ModTime())

		url := filepath.Join(app.conf.BaseURL, fmt.Sprintf("%s/twtxt.txt", name))
		feeds = append(feeds, Feed{name, url, lastModified})
	}

	sort.Slice(feeds, func(i, j int) bool { return feeds[i].Name < feeds[j].Name })

	ctx := struct {
		Title string
		Feeds []Feed
	}{
		Title: "Feeds",
		Feeds: feeds,
	}

	if err := render("feeds", feedsTemplate, ctx, w); err != nil {
		fmt.Fprintf(w, fmt.Errorf("error %w", err).Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
