package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

func main() {
	flag.Parse()

	url := flag.Arg(0)
	filename := flag.Arg(1)

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		log.Fatal(err)
	}

	var lastModified = time.Time{}

	stat, err := os.Stat(filename)
	if err == nil {
		lastModified = stat.ModTime()
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, item := range feed.Items {
		if item.PublishedParsed.After(lastModified) {
			text := fmt.Sprintf("%s\t%s ⌘ %s\n", time.Now().Format(time.RFC3339), item.Title, item.Link)
			n, err := f.WriteString(text)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("appended %d bytes to %s:\n%s", n, filename, text)
		}
	}
}
