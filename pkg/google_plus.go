package collector

import (
  "strconv"
  "regexp"
  "io/ioutil"
  "net/http"
)

func GooglePlus() Platform {
  return Platform{
    enabled:true,
    name:"google_plus",
    statsUrl: "https://plusone.google.com/_/+1/fastbutton?url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      jsBody := string(body)
      count := 0
      matches := regexp.MustCompile("\\s\\{c:\\s(\\d+?)\\.").FindStringSubmatch(jsBody)
      if len(matches) > 1 {
        newCount, error := strconv.Atoi(matches[1]);
        if error != nil {
          return Stat{}, error
        }
        count = newCount
      }

      return Stat{
        data: map[string]interface{}{"count": count},
      }, nil
    },
  }
}
