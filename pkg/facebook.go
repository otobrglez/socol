package collector

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

func Facebook() Platform {
	return Platform{
		enabled:  true,
		name:     "facebook",
		statsURL: "https://api.facebook.com/method/links.getStats?format=json&urls=%s",
		parseWith: func(r *http.Response) (Stat, error) {
			body, error := ioutil.ReadAll(r.Body)
			if error != nil {
				return Stat{}, error
			}

			var jsonBlob []map[string]interface{}
			if err := json.Unmarshal(body, &jsonBlob); err != nil {
				return Stat{}, err
			}

			if len(jsonBlob) < 1 {
				return Stat{}, errors.New("No data")
			}

			keys := []string{"share_count", "like_count", "comment_count", "total_count", "click_count", "commentsbox_count"}
			stat := Stat{data: map[string]interface{}{}}
			for _, key := range keys {
				stat.data[key] = jsonBlob[0][key]
			}

			stat.data["count"] = stat.data["total_count"]
			return stat, nil
		}}
}
