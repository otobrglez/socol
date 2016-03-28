package collector

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
)

func Reddit() Platform {
  return Platform{
    enabled:true,
    name:"reddit",
    statsUrl: "https://www.reddit.com/api/info.json?&url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      //TODO: Implement this
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      ups, downs := 0.0, 0.0
      var jsonBlob map[string]interface{}
      if err := json.Unmarshal(body, &jsonBlob); err != nil {
        return Stat{}, err
      } else {
        pData := jsonBlob["data"].(map[string]interface{})
        for _, ch := range pData["children"].([]interface{}) {
          el := ch.(map[string]interface{})
          if el["kind"] == "t3" {
            d := el["data"].(map[string]interface{})
            ups += d["ups"].(float64)
            downs += d["downs"].(float64)
          }
        }
      }

      return Stat{
        data: map[string]interface{}{
          "ups": ups,
          "downs": downs,
          "count": ups + downs,
        },
      }, nil
    },
  }
}
