# socol

Social metrics collector written in Go.

[![Build Status][travis-ci-badge]][travis-ci]
[![ImageLayers][imagelayers-badge]][imagelayers]
[![Docker Pulls][docker-pulls-badge]][docker-hub]
[![Docker Stars][docker-stars-badge]][docker-hub]

## Install

```
go get github.com/otobrglez/socol
```

## Usage

```
socol -url https://golang.org/,http://www.scala-lang.org/ -platform facebook,linkedin
```

## Runing as server

```
# Start it on port 6000
socol -s -p 6000
```

This app is ready to be used with [Heroku](https://heroku.com).

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

## Platforms

[socol][socol] supported collection of following metrics

```
https://api.facebook.com/method/links.getStats?format=json&urls=%s
http://api.pinterest.com/v1/urls/count.json?callback=call&url=%s
http://www.linkedin.com/countserv/count/share?url=%s
https://plusone.google.com/_/+1/fastbutton?url=%s
https://www.reddit.com/api/info.json?&url=%s
https://api.bufferapp.com/1/links/shares.json?url=%s
http://www.stumbleupon.com/services/1.01/badge.getinfo?url=%s
```

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

