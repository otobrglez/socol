package collector

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Bufferapp() Platform {
	return Platform{
		enabled:  true,
		name:     "buffer",
		statsURL: "https://api.bufferapp.com/1/links/shares.json?url=%s",
		parseWith: func(r *http.Response) (Stat, error) {
			body, error := ioutil.ReadAll(r.Body)
			if error != nil {
				return Stat{}, error
			}

			var jsonBlob map[string]interface{}
			if err := json.Unmarshal(body, &jsonBlob); err != nil {
				return Stat{}, err
			}

			return Stat{
				data: map[string]interface{}{"count": jsonBlob["shares"]},
			}, nil
		},
	}
}
