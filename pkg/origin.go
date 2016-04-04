package collector

import (
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/fatih/structs"
	"net/http"
	"reflect"
)

func Origin() Platform {
	return Platform{
		enabled:  true,
		name:     "origin",
		statsURL: "%s",
		parseWith: func(r *http.Response) (Stat, error) {
			og := opengraph.NewOpenGraph()
			err := og.ProcessHTML(r.Body)
			if err != nil {
				return Stat{}, err
			}

			data := map[string]interface{}{}
			asMap := structs.Map(og)
			for k, v := range asMap {
				if v != nil && v != "" {
					val := reflect.ValueOf(v)
					if val.Kind() == reflect.Ptr || val.Kind() == reflect.Slice || val.Kind() == reflect.Map {
						// TODO: Do something if its pointer
					} else {
						data[k] = v
					}
				}
			}

			return Stat{data: data}, nil
		},
	}
}
