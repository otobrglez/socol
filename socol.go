package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/otobrglez/socol/pkg"
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

	aggregated := collector.New(url, platforms, proxy)

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

var logger *log.Logger
var errorsLogger *log.Logger
var isServer = false
var cliURLs = []string{}
var cliURL = ""
var cliPlatforms = []string{}
var cliPlatform = ""
var port = 5000
var proxy = ""

func init() {
	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}

	logger = log.New(os.Stdout, "socol-cmd ", log.Ldate|log.Ltime|log.Lshortfile)
	errorsLogger = log.New(os.Stderr, "socol-cmd ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	flag.BoolVar(&isServer, "s", false, "run as server")
	flag.StringVar(&cliURL, "url", "", "url(s) to fetch")
	flag.StringVar(&cliPlatform, "platform", "", "platform(s) to fetch")
	flag.IntVar(&port, "p", 5000, "server port")
	flag.StringVar(&proxy, "proxy", "", "proxy")

	proxyEnv := os.Getenv("PROXY")
	if proxy == "" && proxyEnv != "" {
		proxy = proxyEnv
	}

	proxyEnv = os.Getenv("PROXY_MESH")
	if proxyEnv != "" {
		proxy = proxyEnv
	}

	flag.Parse()

	if !isServer {
		cliURLs := strings.Split(cliURL, ",")
		cliPlatforms := strings.Split(cliPlatform, ",")

		for _, url := range cliURLs {
			if len(cliPlatforms) == 0 ||
				(len(cliPlatforms) == 1 && cliPlatforms[0] == "") {
				cliPlatforms = nil
			}

			aggregated := collector.New(url, cliPlatforms, proxy)

			body, error := json.MarshalIndent(aggregated, "", "  ")
			if error != nil {
				panic(error)
			}

			fmt.Println(string(body))
		}

		os.Exit(0)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("socol."))
	})

	http.HandleFunc("/stats", statsHandler)

	portAsString := os.Getenv("PORT")
	if portAsString != "" {
		port, _ = strconv.Atoi(portAsString)
	}

	logger.Println("Listening on", port)
	error := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if error != nil {
		errorsLogger.Fatal("Error listening on ", port)
		os.Exit(2)
	} else {
		logger.Println("Started.")
	}
}
