package collector

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
)

func Stumbleupon() Platform {
	return Platform{
		enabled:  true,
		name:     "stumbleupon",
		statsURL: "http://www.stumbleupon.com/services/1.01/badge.getinfo?url=%s",
		parseWith: func(r *http.Response) (Stat, error) {
			body, error := ioutil.ReadAll(r.Body)
			if error != nil {
				return Stat{}, error
			}

			var jsonBlob map[string]interface{}
			if err := json.Unmarshal(body, &jsonBlob); err != nil {
				return Stat{}, err
			}

			result := jsonBlob["result"].(map[string]interface{})
			count := float64(0)

			if result["in_index"] == true {
				p := reflect.TypeOf(result["views"])
				if p.Kind() == reflect.String {
					vInt, _ := strconv.Atoi(result["views"].(string))
					count = float64(vInt)
				} else if p.Kind() == reflect.Float64 {
					count = float64(result["views"].(float64))
				} else {
					panic("stumbleupon - no idea ;/")
				}
			}

			return Stat{
				data: map[string]interface{}{"count": count},
			}, nil
		},
	}
}
