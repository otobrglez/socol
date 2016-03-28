package collector

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
)

func Linkedin() Platform {
  return Platform{
    enabled:true,
    name:"linkedin",
    statsUrl: "http://www.linkedin.com/countserv/count/share?url=%s",
    format: "jsonp",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      jsonBody, error := parseJSONP(body)
      if error != nil {
        return Stat{}, error
      }

      var jsonBlob map[string]interface{}
      if err := json.Unmarshal([]byte(jsonBody), &jsonBlob); err != nil {
        return Stat{}, err
      }

      return Stat{
        data: map[string]interface{}{"count": jsonBlob["count"]},
      }, nil
    },
  }
}
