package collector

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

func Pocket() Platform {
	return Platform{
		enabled:  true,
		name:     "pocket",
		statsURL: "https://widgets.getpocket.com/v1/button?count=vertical&url=%s",
		parseWith: func(r *http.Response) (Stat, error) {
			body, error := ioutil.ReadAll(r.Body)
			if error != nil {
				return Stat{}, error
			}

			jsBody := string(body)
			count := 0
			matches := regexp.MustCompile("\\sid=\"cnt\">(\\d+)</em>").FindStringSubmatch(jsBody)
			if len(matches) > 1 {
				newCount, error := strconv.Atoi(matches[1])
				if error != nil {
					return Stat{}, error
				}
				count = newCount
			}

			return Stat{
				data: map[string]interface{}{
					"count": count,
				},
			}, nil
		},
	}
}
