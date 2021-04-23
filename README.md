# rss2twtxt

[![Build Status](https://cloud.drone.io/api/badges/prologic/rss2twtxt/status.svg)](https://cloud.drone.io/prologic/rss2twtxt)
[![CodeCov](https://codecov.io/gh/prologic/rss2twtxt/branch/master/graph/badge.svg)](https://codecov.io/gh/prologic/rss2twtxt)
[![Go Report Card](https://goreportcard.com/badge/prologic/rss2twtxt)](https://goreportcard.com/report/prologic/rss2twtxt)
[![GoDoc](https://godoc.org/github.com/prologic/rss2twtxt?status.svg)](https://godoc.org/github.com/prologic/rss2twtxt) 
[![Sourcegraph](https://sourcegraph.com/github.com/prologic/rss2twtxt/-/badge.svg)](https://sourcegraph.com/github.com/prologic/rss2twtxt?badge)

`rss2twtxt` is an RSS/Atom feed aggregator for [twtxt](https://rss2twtxt.readthedocs.io/en/latest/)
that consumes RSS/Atom feeds and processes them into twtxt feeds. These can
then be consumed by any standard twtxt client such as:

- [twtxt](https://github.com/buckket/twtxt)
- [twet](https://github.com/quite/twet)
- [txtnish](https://github.com/mdom/txtnish)
- [twtxtc](https://github.com/neauoire/twtxtc)

There is also a publically (_free_) service online available at:

- https://feeds.twtxt.cc/

![Screenshot 1](./screenshot1.png)
![Screenshot 2](./screenshot2.png)

## Installation

### Source

```#!bash
$ go get -u github.com/twtpub/rss2twtxt
```

## Usage

Run `rss2twtxt`:

```#!bash
$ rss2twtxt
```

Then visit: http://localhost:8000/

## Related Projects

- [twtxt](https://github.com/twtpub/twtxt)

## License

`rss2twtxt` is licensed under the terms of the [MIT License](/LICENSE)
