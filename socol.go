package socol

/*
  - https://github.com/abeMedia/shareCount/blob/master/share_count.php
*/

import (
  "fmt"
  "reflect"
  "net/http"
  "log"
  "io/ioutil"
  "encoding/json"
  "errors"
  "time"
  "strings"
  "regexp"
  "strconv"
  _ "os"
)

type Stat struct {
  name string
  data map[string]interface{}
}

type Platform struct {
  enabled   bool
  name      string
  statsUrl  string
  parseWith func(*http.Response) (Stat, error)
  stat      Stat
}

var platforms []Platform = []Platform{
  Platform{
    enabled:true,
    name:"facebook",
    statsUrl: "https://api.facebook.com/method/links.getStats?format=json&urls=%s",
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
    }},
  Platform{
    enabled:true,
    name:"pinterest",
    statsUrl: "http://api.pinterest.com/v1/urls/count.json?callback=call&url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      jsBody := string(body)
      jsonBody := jsBody[strings.Index(jsBody, "(") + 1:strings.LastIndex(jsBody, ")")]

      var jsonBlob map[string]interface{}
      if err := json.Unmarshal([]byte(jsonBody), &jsonBlob); err != nil {
        return Stat{}, err
      }

      return Stat{
        data: map[string]interface{}{"count": jsonBlob["count"]},
      }, nil
    }},
  Platform{
    enabled:true,
    name:"linkedin",
    statsUrl: "http://www.linkedin.com/countserv/count/share?url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      jsBody := string(body)
      jsonBody := jsBody[strings.Index(jsBody, "(") + 1:strings.LastIndex(jsBody, ")")]

      var jsonBlob map[string]interface{}
      if err := json.Unmarshal([]byte(jsonBody), &jsonBlob); err != nil {
        return Stat{}, err
      }

      return Stat{
        data: map[string]interface{}{"count": jsonBlob["count"]},
      }, nil
    },
  },
  Platform{
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
  },
  Platform{
    enabled:false,
    name:"reddit",
    statsUrl: "https://www.reddit.com/api/info.json?&url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      fmt.Println(string(body))

      var jsonBlob map[string]interface{}
      if err := json.Unmarshal(body, &jsonBlob); err != nil {
        return Stat{}, err
      }

      fmt.Println(jsonBlob["data"])

      return Stat{
      }, nil
    },
  },
}

func (platform Platform) doRequest(lookupUrl string, stats chan <- Stat, errors chan *error) {
  start := time.Now()
  fullUrl := fmt.Sprintf(platform.statsUrl, lookupUrl)
  logger.Println(platform.name, "Requesting", fullUrl)

  response, error := http.Get(fullUrl)
  if error != nil {
    errors <- &error
    return
  }
  fetchedIn := time.Now().Sub(start).Seconds()

  stat, error := platform.parseWith(response)
  if error != nil {
    errors <- &error
    return
  }

  if stat.data == nil {
    stat.data = map[string]interface{}{}
  }

  stat.data["fetched_in"] = fetchedIn
  stat.data["completed_in"] = time.Now().Sub(start).Seconds()

  logger.Println(platform.name, "Completed in", stat.data["completed_in"], "s")

  stat.name = platform.name
  stats <- stat
  return
}

var logger *log.Logger

func CollectStats(lookupUrl string) (map[string]interface{}) {

  logger = log.New(ioutil.Discard, "socol", log.Ldate | log.Ltime | log.Lshortfile)
  errors, stats, taskCount := make(chan *error), make(chan Stat), 0

  for _, platform := range platforms {
    if platform.enabled {
      go platform.doRequest(lookupUrl, stats, errors)
      taskCount++
    }
  }

  aggregated := map[string]interface{}{}

  for {
    select {
    case stat := <-stats:
      aggregated[stat.name] = stat.data
      taskCount--
    case e := <-errors:
      logger.Fatal("ERROR ~~> ", (*e))
      taskCount--
    default:
      if taskCount <= 0 {
        total := 0
        for _, p := range aggregated {
          c := p.(map[string]interface{})["count"]
          if reflect.ValueOf(c).Kind() == reflect.Int {
            total += c.(int)
          } else if reflect.ValueOf(c).Kind() == reflect.Float64 {
            n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 64));
            total += n
          } else if reflect.ValueOf(c).Kind() == reflect.Float32 {
            n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 32));
            total += n
          }else {
            logger.Fatal("Can't cast...")
          }
        }

        aggregated["total"] = map[string]interface{}{"total": total}

        return aggregated
      }
    }
  }
}

func CollectStatsFor(lookupUrl string, sites interface{}) {
  fmt.Println(sites)
  fmt.Println(reflect.TypeOf(sites))
  fmt.Println(reflect.ValueOf(sites).Kind())
}
