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

const (
	Msg500 = "内部服务器错误"
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
	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx := struct {
			Title string
		}{
			Title: "RSS2Twt 中文版",
		}

		if r.Method == http.MethodHead {
			return
		}

		if err := render("index", indexTemplate, ctx, w); err != nil {
			log.WithError(err).Error("error rending index template")
			http.Error(w, Msg500, http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		url := r.FormValue("url")

		if url == "" {
			if err := renderMessage(w, http.StatusBadRequest, "错误", "没有URL参数"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, Msg500, http.StatusInternalServerError)
			}
			return
		}

		feed, err := ValidateFeed(app.conf, url)
		if err != nil {
			if err := renderMessage(w, http.StatusBadRequest, "错误", fmt.Sprintf("不能找到有效的 RSS/Atom 源: %s", url)); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, Msg500, http.StatusInternalServerError)
			}
			return
		}

		if _, ok := app.conf.Feeds[feed.Name]; ok {
			if err := renderMessage(w, http.StatusConflict, "错误", "Feed 源已经存在"); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, Msg500, http.StatusInternalServerError)
			}
			return
		}

		app.conf.Feeds[feed.Name] = feed.URL
		if err := app.conf.Save(); err != nil {
			msg := fmt.Sprintf("不能保存 Feed: %s", err)
			if err := renderMessage(w, http.StatusInternalServerError, "错误", msg); err != nil {
				log.WithError(err).Error("error rendering message template")
				http.Error(w, Msg500, http.StatusInternalServerError)
			}
			return
		}

		msg := fmt.Sprintf("添加 [%s](%s) Feed 源成功", feed.Name, feed.URL)
		if err := renderMessage(w, http.StatusCreated, "成功", msg); err != nil {
			log.WithError(err).Error("error rendering message template")
			http.Error(w, Msg500, http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}

func (app *App) FeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		vars := mux.Vars(r)

		name := vars["name"]
		if name == "" {
			http.Error(w, "错误请求", http.StatusBadRequest)
			return
		}

		filename := filepath.Join(app.conf.Root, fmt.Sprintf("%s.txt", name))
		if !Exists(filename) {
			log.Warnf("feed does not exist %s", name)
			http.Error(w, "Feed 没有找到", http.StatusNotFound)
			return
		}

		fileInfo, err := os.Stat(filename)
		if err != nil {
			log.WithError(err).Error("os.Stat() error")
			http.Error(w, Msg500, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		if r.Method == http.MethodHead {
			return
		}

		http.ServeFile(w, r, filename)
		return
	}
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}

func (app *App) AvatarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, no-cache, must-revalidate")

		vars := mux.Vars(r)

		name := vars["name"]
		if name == "" {
			http.Error(w, "错误请求", http.StatusBadRequest)
			return
		}

		filename := filepath.Join(app.conf.Root, fmt.Sprintf("%s.txt", name))
		if !Exists(filename) {
			log.Warnf("feed does not exist %s", name)
			http.Error(w, "Feed 没有找到", http.StatusNotFound)
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
				http.Error(w, Msg500, http.StatusInternalServerError)
				return
			}
			defer f.Close()

			fileInfo, err := os.Stat(fn)
			if err != nil {
				log.WithError(err).Error("os.Stat() error")
				http.Error(w, Msg500, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

			if r.Method == http.MethodHead {
				return
			}

			if _, err := io.Copy(w, f); err != nil {
				log.WithError(err).Error("error writing avatar response")
				http.Error(w, Msg500, http.StatusInternalServerError)
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

		buf := bytes.Buffer{}
		img := cameron.Identicon([]byte(name), avatarResolution, 12)
		png.Encode(&buf, img)

		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

		if r.Method == http.MethodHead {
			return
		}

		w.Write(buf.Bytes())
		return
	}
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}

func (app *App) WeAreFeedsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if r.Method == http.MethodHead {
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, feed := range app.GetFeeds() {
			fmt.Fprintf(w, "%s %s\n", feed.Name, feed.URL)
		}
		return
	}
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}

func (app *App) FeedsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		if accept.PreferredContentTypeLike(r.Header, "text/plain") == "text/plain" {
			app.WeAreFeedsHandler(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx := struct {
			Title string
			Feeds []Feed
		}{
			Title: "可用的 Twtxt Feed 源",
			Feeds: app.GetFeeds(),
		}

		if r.Method == http.MethodHead {
			return
		}

		if err := render("feeds", feedsTemplate, ctx, w); err != nil {
			log.WithError(err).Error("error rendering feeds template")
			http.Error(w, Msg500, http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}
