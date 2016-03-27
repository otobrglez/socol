package collector

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
  format    string
}

var Formats = map[string]string{
  "xml": "text/xml",
  "jsonp": "application/javascript",
  "json": "application/json",
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
    }},
  Platform{
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
    enabled:true,
    name:"pocket",
    statsUrl:"https://widgets.getpocket.com/v1/button?count=vertical&url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      jsBody := string(body)
      count := 0
      matches := regexp.MustCompile("\\sid=\"cnt\">(\\d+)</em>").FindStringSubmatch(jsBody)
      if len(matches) > 1 {
        newCount, error := strconv.Atoi(matches[1]);
        if error != nil {
          return Stat{}, error
        }
        count = newCount
      }

      return Stat{
        data: map[string]interface{}{
          "count": count,
        },
      }, nil

      return Stat{}, nil
    },
  },
  Platform{
    enabled:true,
    name: "tumblr",
    format: "json",
    statsUrl: "http://api.tumblr.com/v2/share/stats?url=%s",
    parseWith: func(r *http.Response) (Stat, error) {
      body, error := ioutil.ReadAll(r.Body)
      if error != nil {
        return Stat{}, error
      }

      count := 0.0
      var jsonBlob map[string]interface{}
      if err := json.Unmarshal(body, &jsonBlob); err != nil {
        return Stat{}, err
      } else {
        count = jsonBlob["response"].(map[string]interface{})["note_count"].(float64)
      }

      return Stat{
        data: map[string]interface{}{"count": count, },
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

func parseJSONP(body []byte) (string, error) {
  jsBody := string(body)
  iStart := strings.Index(jsBody, "(")
  iEnd := strings.LastIndex(jsBody, ")")

  if iStart == -1 || iEnd == -1 {
    return "", errors.New("Something is wrong with payload.")
  }

  return jsBody[iStart + 1:iEnd], nil
}

func buildClientAsync() (*http.Client, error) {
  transport := &http.Transport{
    TLSClientConfig: &tls.Config{
      InsecureSkipVerify: true,
    },
  }

  if proxy != "" {
    proxyThing, error := url.Parse(proxy)
    if error != nil {
      return &http.Client{}, error
    }

    transport.Proxy = http.ProxyURL(proxyThing)
  }

  return &http.Client{
    Timeout: globalTimeout,
    Transport: transport,
  }, nil
}

func (platform Platform) doRequest(lookupUrl string, stats chan <- Stat, errorsChannel chan *error) {
  start := time.Now()
  fullUrl := fmt.Sprintf(platform.statsUrl, lookupUrl)
  logger.Println(platform.name, "Requesting", fullUrl)

  client, err := buildClientAsync()
  if err != nil {
    errorsChannel <- &err
    return
  }

  request, error := http.NewRequest("GET", fullUrl, nil)
  request.Header.Set("User-Agent", strings.Join([]string{"Mozilla/5.0 (socol) ", strconv.Itoa(rand.Intn(1000))}, " "))
  if platform.format != "" {
    logger.Println("Setting content type to", platform.format)
    request.Header.Set("Content-Type", platform.format)
  }

  // request.Header.Set("X-Requested-With", "XMLHttpRequest")

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

func resolveAndOG(url string) (stat Stat, urls []string, err error) {
  start := time.Now()
  stat.name = "origin"
  err = nil
  urls = append(urls, url)

  client, e := buildClientAsync()
  if e != nil {
    err = e
    return
  }

  logger.Println(stat.name, "Requesting", url)

  client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
    urls = append(urls, req.URL.String())
    logger.Println(stat.name, "Got", req.URL.String())
    return nil
  }

  request, e := http.NewRequest("GET", url, nil)
  request.Header.Set("User-Agent", strings.Join([]string{"Mozilla/5.0 (socol) ", strconv.Itoa(rand.Intn(1000))}, " "))
  if e != nil {
    err = e
    return
  }

  response, e := client.Do(request)
  if e != nil {
    err = e
    return
  }

  if response.StatusCode != http.StatusOK {
    err = errors.New("Got non OK HTTP status at " + response.Status + "-" + url)
    return
  }

  fetchedIn := time.Now().Sub(start).Seconds()
  stat, e = platforms[len(platforms) - 1].parseWith(response)
  if e != nil {
    err = e
    return
  }

  stat.name = "origin"
  stat.data["fetched_in"] = fetchedIn
  stat.data["completed_in"] = time.Now().Sub(start).Seconds()
  logger.Println(stat.name, "Completed in", stat.data["completed_in"], "s")

  if stat.data == nil {
    stat.data = map[string]interface{}{}
  } else {
    stat.data["urls"] = urls
  }

  return
}

func canRunPlatform(platform *Platform, selectedPlatforms *[]string) (canRun bool) {
  canRun = false
  if platform.name == "origin" {
    return false
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

var proxy = ""
var logger *log.Logger
var errorsLogger *log.Logger
var globalTimeout = time.Duration(4 * time.Second)

func init() {
  logLevel := os.Getenv("LOG_LEVEL")
  if logLevel == "" {
    logger = log.New(ioutil.Discard, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
    errorsLogger = log.New(ioutil.Discard, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
  } else {
    logger = log.New(os.Stdout, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
    errorsLogger = log.New(os.Stderr, "socol ", log.Ldate | log.Ltime | log.Lshortfile)
  }
}

func New(lookupUrl string, selectedPlatforms []string, pproxy string) (map[string]interface{}) {
  proxy = pproxy

  if selectedPlatforms == nil ||
  (len(selectedPlatforms) == 1 && selectedPlatforms[0] == "") {
    selectedPlatforms = []string{}
  }

  selectedPlatforms = append(selectedPlatforms, "origin")
  errors, stats, taskCount := make(chan *error), make(chan Stat), 0
  aggregated := map[string]interface{}{}

  rStat, urls, rError := resolveAndOG(lookupUrl)
  if rError != nil {
    errorsLogger.Println(rError)
  } else {
    aggregated[rStat.name] = rStat.data
  }

  if len(urls) > 1 {
    logger.Println("Digging for", urls)
  }

  lookupUrl = urls[len(urls) - 1]

  for _, platform := range platforms {
    if canRunPlatform(&platform, &selectedPlatforms) {
      go platform.doRequest(lookupUrl, stats, errors)
      taskCount++
    }
  }

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
