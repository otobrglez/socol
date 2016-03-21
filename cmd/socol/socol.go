package main

import (
  "fmt"
  "runtime"
  "os"
  socol "github.com/otobrglez/socol"
  "strings"
  "encoding/json"
)

func init() {
  if cpu := runtime.NumCPU(); cpu == 1 {
    runtime.GOMAXPROCS(2)
  } else {
    runtime.GOMAXPROCS(cpu)
  }
}

func main() {
  if len(os.Args) < 2 {
    panic("Missing lookup URL as first argument!")
  }

  urls := strings.Split(os.Args[1], ",")

  platforms := []string{}
  if len(os.Args) > 2 {
    platforms = strings.Split(os.Args[2], ",")
  }

  for _, url := range urls {
    aggregated := socol.CollectStats(url, platforms)

    body, error := json.Marshal(aggregated)
    if error != nil {
      panic(error)
    }

    fmt.Println(string(body))
  }
}
