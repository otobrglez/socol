package main

import (
  "net/http"
  "log"
  "os"
  "io/ioutil"
  "strconv"
  "encoding/json"
  "runtime"
  "github.com/otobrglez/socol"
  "strings"
)

func statsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  query := r.URL.Query()
  url := query.Get("url")

  if url == "" {
    json, _ := json.Marshal(map[string]interface{}{"error": "Missing url."})
    http.Error(w, string(json), http.StatusBadRequest)
    return
  }

  platforms := strings.Split(query.Get("platforms"), ",")
  if len(platforms) == 1 && platforms[0] == "" {
    platforms = nil
  }

  aggregated := socol.CollectStats(url, platforms)

  body, error := json.Marshal(aggregated)
  if error != nil {
    http.Error(w, error.Error(), http.StatusInternalServerError)
    return
  }

  w.Write(body)
}

var logger *log.Logger

func init() {
  if cpu := runtime.NumCPU(); cpu == 1 {
    runtime.GOMAXPROCS(2)
  } else {
    runtime.GOMAXPROCS(cpu)
  }
}

func main() {
  logLevel := os.Getenv("LOG_LEVEL")
  if logLevel == "" {
    logger = log.New(ioutil.Discard, "server ", log.Ldate | log.Ltime | log.Lshortfile)
  } else {
    logger = log.New(os.Stdout, "server ", log.Ldate | log.Ltime | log.Lshortfile)
  }

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("socol."))
  })

  http.HandleFunc("/stats", statsHandler)

  port := 8080
  logger.Println("Starting server on", port)
  error := http.ListenAndServe(":" + strconv.Itoa(port), nil)
  if error != nil {
    logger.Fatal("Error starting!", port)
    os.Exit(2)
  } else {
    logger.Println("Started.")
  }

}
