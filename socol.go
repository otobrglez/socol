package socol

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
  "math/rand"
  "github.com/fatih/structs"
  "github.com/dyatlov/go-opengraph/opengraph"
  "os"
  "crypto/tls"
  "net/url"
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
      iStart := strings.Index(jsBody, "(")
      iEnd := strings.LastIndex(jsBody, ")")

      if iStart == -1 || iEnd == -1 {
        error := errors.New("Something is wrong with payload.")
        return Stat{}, error
      }

      jsonBody := jsBody[iStart + 1:iEnd]

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

      iStart := strings.Index(jsBody, "(")
      iEnd := strings.LastIndex(jsBody, ")")

      if iStart == -1 || iEnd == -1 {
        error := errors.New("Something is wrong with payload.")
        return Stat{}, error
      }

      jsonBody := jsBody[iStart + 1:iEnd]

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
      //TODO: Implement this
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
  Platform{
    enabled: true,
    name: "buffer",
    statsUrl: "https://api.bufferapp.com/1/links/shares.json?url=%s",
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
  },
  Platform{
    enabled: true,
    name: "stumbleupon",
    statsUrl: "http://www.stumbleupon.com/services/1.01/badge.getinfo?url=%s",
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
  },
  Platform{
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
  },
}

func (platform Platform) doRequest(lookupUrl string, proxy string, stats chan <- Stat, errorsChannel chan *error) {
  start := time.Now()
  fullUrl := fmt.Sprintf(platform.statsUrl, lookupUrl)
  logger.Println(platform.name, "Requesting", fullUrl)

  transport := &http.Transport{
    TLSClientConfig: &tls.Config{
      InsecureSkipVerify: true,
    },
  }

  if proxy != "" {
    proxyThing, error := url.Parse(proxy)
    if error != nil {
      errorsChannel <- &error
      return
    }

    transport.Proxy = http.ProxyURL(proxyThing)
  }

  client := &http.Client{
    Timeout: time.Duration(3 * time.Second),
    Transport: transport,
  }

  request, error := http.NewRequest("GET", fullUrl, nil)
  request.Header.Set("User-Agent", strings.Join([]string{"Mozilla/5.0 (socol) ", strconv.Itoa(rand.Intn(1000))}, " "))
  if error != nil {
    errorsChannel <- &error
    return
  }

  response, error := client.Do(request)
  if error != nil {
    errorsChannel <- &error
    return
  }

  if response.StatusCode != http.StatusOK {
    error := errors.New("Got non OK HTTP status at " + response.Status + "-" + fullUrl)
    errorsChannel <- &error
  }

  fetchedIn := time.Now().Sub(start).Seconds()

  stat, error := platform.parseWith(response)
  if error != nil {
    errorsChannel <- &error
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
var errorsLogger *log.Logger

func CollectStats(lookupUrl string, selectedPlatforms []string, proxy string) (map[string]interface{}) {
  if selectedPlatforms == nil ||
  (len(selectedPlatforms) == 1 && selectedPlatforms[0] == "") {
    selectedPlatforms = []string{}
  }

  selectedPlatforms = append(selectedPlatforms, "origin")

  logLevel := os.Getenv("LOG_LEVEL")
  if logLevel == "" {
    logger = log.New(ioutil.Discard, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
    errorsLogger = log.New(ioutil.Discard, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
  } else {
    logger = log.New(os.Stdout, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
    errorsLogger = log.New(os.Stderr, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
  }

  errors, stats, taskCount := make(chan *error), make(chan Stat), 0

  for _, platform := range platforms {
    if canRunPlatform(&platform, &selectedPlatforms) {
      go platform.doRequest(lookupUrl, proxy, stats, errors)
      taskCount++
    }
  }

  aggregated := map[string]interface{}{}

  for {
    select {
    case stat := <-stats:
      aggregated[stat.name] = stat.data
      taskCount--
    case error := <-errors:
      errorsLogger.Println(*error)
      taskCount--
    default:
      if taskCount <= 0 {
        return aggregateAndCombine(aggregated)
      }
    }
  }
}

func canRunPlatform(platform *Platform, selectedPlatforms *[]string) (canRun bool) {
  canRun = false
  if platform.name == "origin" {
    return true
  }

  if platform.enabled == false {
    return false
  }

  if len(*selectedPlatforms) == 1 {
    return true
  }

  for _, name := range *selectedPlatforms {
    if platform.name == name && platform.enabled == true {
      canRun = true
      return true
    }
  }

  return
}

func aggregateAndCombine(results map[string]interface{}) (map[string]interface{}) {
  total := 0
  for _, p := range results {
    c := p.(map[string]interface{})["count"]
    if reflect.ValueOf(c).Kind() == reflect.Int {
      total += c.(int)
    } else if reflect.ValueOf(c).Kind() == reflect.Float64 {
      n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 64));
      total += n
    } else if reflect.ValueOf(c).Kind() == reflect.Float32 {
      n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 32));
      total += n
    } else {
      // logger.Fatal("Can't cast...")
    }
  }

  results["meta"] = map[string]interface{}{"total": total}
  return results
}
