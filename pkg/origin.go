package collector

import (
  "reflect"
  "github.com/fatih/structs"
  "github.com/dyatlov/go-opengraph/opengraph"
  "net/http"
)

func Origin() Platform {
  return Platform{
    enabled: true,
    name: "origin",
    statsUrl: "%s",
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

      return Stat{data: data, }, nil
    },
  }
}
