package main

import (
  "net/http"
  "log"
  "os"
  "strconv"
  "encoding/json"
  "runtime"
  "github.com/otobrglez/socol"
  "strings"
  "time"
)

func statsHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  start := time.Now()
  query := r.URL.Query()
  url := query.Get("url")

  if url == "" {
    error := "Missing required URL."
    json, _ := json.Marshal(map[string]interface{}{"error": error})
    http.Error(w, string(json), http.StatusBadRequest)
    errorsLogger.Println("Failed", error)
    return
  }

  platforms := strings.Split(query.Get("platforms"), ",")
  if len(platforms) == 1 && platforms[0] == "" {
    platforms = nil
  }

  aggregated := socol.CollectStats(url, platforms)

  body, error := json.Marshal(aggregated)
  if error != nil {
    error := "Error compiling JSON."
    json, _ := json.Marshal(map[string]interface{}{"error": error})
    http.Error(w, string(json), http.StatusInternalServerError)
    return
  }

  logger.Println("Compiled stats for", url, "in", time.Now().Sub(start).Seconds(), "sec.")
  w.Write(body)
}

func init() {
  if cpu := runtime.NumCPU(); cpu == 1 {
    runtime.GOMAXPROCS(2)
  } else {
    runtime.GOMAXPROCS(cpu)
  }
}

var logger *log.Logger
var errorsLogger *log.Logger

func main() {
  logger = log.New(os.Stdout, "socol-server ", log.Ldate | log.Ltime | log.Lshortfile)
  errorsLogger = log.New(os.Stderr, "socol-server ", log.Ldate | log.Ltime | log.Lshortfile)

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("socol."))
  })

  http.HandleFunc("/stats", statsHandler)

  portAsString := os.Getenv("PORT")
  port := 5000
  if portAsString != "" {
    port, _ = strconv.Atoi(portAsString)
  }

  logger.Println("socol-server on", port)
  error := http.ListenAndServe(":" + strconv.Itoa(port), nil)
  if error != nil {
    errorsLogger.Fatal("Error starting on ", port)
    os.Exit(2)
  } else {
    logger.Println("Started.")
  }

}
