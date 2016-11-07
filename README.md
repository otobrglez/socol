# socol

Social metrics collector written in Go.

[![Build Status][travis-ci-badge]][travis-ci]
[![Go Report Card][goreportcard-badge]][goreportcard]
[![ImageLayers][imagelayers-badge]][imagelayers]
[![Docker Pulls][docker-pulls-badge]][docker-hub]
[![Docker Stars][docker-stars-badge]][docker-hub]

## Install

```
go get github.com/otobrglez/socol
```

## Usage

Collect stats for one URL.

```
socol -url https://golang.org/
```

Collect stats for multiple URLs and selected platforms.
```
socol -url https://golang.org/,http://www.scala-lang.org/ -platform facebook,linkedin
```

## Running as server

Start it on port 6000.

```
socol -s -p 6000

# Try it,...
curl "http://127.0.0.1:6000/stats?url=https://golang.org/"
```

This app is ready to be used with [Heroku](https://heroku.com) or [Docker (instructions)](#docker).

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

## Platforms

[socol][socol] supported collection of following metrics: [Buffer](https://buffer.com/), [Facebook](http://fb.com),
[Google Plus](https://plus.google.com/), [LinkedIn](https://www.linkedin.com/), [Pinterest](https://www.pinterest.com/), [Pocket](https://getpocket.com), [Reddit](https://www.reddit.com), [StumbleUpon](https://www.stumbleupon.com/), [Tumblr](https://www.tumblr.com/).

> Why is [Twitter](https://twitter.com/) not supported? Twitter has decided to remove stats from their public interfaces. You can read more about [why on their blog](https://blog.twitter.com/2015/hard-decisions-for-a-sustainable-platform).

## Docker

Running [socol][socol] server via Docker.

```bash
docker run -ti -p 5000:5000 otobrglez/socol

curl -s <docker_host>:5000/stats\?url=http://www.facebook.com | python -mjson.tool
```

Response from service.

```json
{
    "facebook": {
        "click_count": 0,
        "comment_count": 7837699.0,
        "commentsbox_count": 5177,
        "completed_in": 0.28217279100000003,
        "count": 57893662.0,
        "fetched_in": 0.282076959,
        "like_count": 20070293.0,
        "share_count": 29985670.0,
        "total_count": 57893662.0
    },
    "google_plus": {
        "completed_in": 0.31490062700000004,
        "count": 338527,
        "fetched_in": 0.269049992
    },
    "linkedin": {
        "completed_in": 0.78390614,
        "count": 4479,
        "fetched_in": 0.783683243
    },
    "meta": {
        "total": 58320895
    },
    "origin": {
        "Locale": "sl_SI",
        "SiteName": "Facebook",
        "URL": "https://www.facebook.com/",
        "completed_in": 0.343662642,
        "fetched_in": 0.33344814500000003
    },
    "pinterest": {
        "completed_in": 0.143030286,
        "count": 60256,
        "fetched_in": 0.142972456
    },
    "stumbleupon": {
        "completed_in": 0.435604561,
        "count": 23971,
        "fetched_in": 0.435491601
    }
}
```

## Author

- [Oto Brglez][me]

## License

Use it under `MIT`.

[socol]: https://github.com/otobrglez/socol
[me]: https://github.com/otobrglez
[travis-ci]: https://travis-ci.org/otobrglez/socol
[travis-ci-badge]: https://travis-ci.org/otobrglez/socol.svg?branch=master
[imagelayers-badge]: https://badge.imagelayers.io/otobrglez/socol:latest.svg
[imagelayers]: https://imagelayers.io/?images=otobrglez/socol:latest
[docker-pulls-badge]: https://img.shields.io/docker/pulls/otobrglez/socol.svg
[docker-stars-badge]: https://img.shields.io/docker/stars/otobrglez/socol.svg
[docker-hub]: https://hub.docker.com/r/otobrglez/socol/             
[goreportcard-badge]: https://goreportcard.com/badge/otobrglez/socol
[goreportcard]: https://goreportcard.com/report/otobrglez/socol
