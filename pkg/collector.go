package collector

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	"xml":   "text/xml",
	"jsonp": "application/javascript",
	"json":  "application/json",
}

var platforms []Platform = []Platform{
	Facebook(),
	Pinterest(),
	Linkedin(),
	GooglePlus(),
	Reddit(),
	Bufferapp(),
	Stumbleupon(),
	Pocket(),
	Tumblr(),
	Origin(),
}

func parseJSONP(body []byte) (string, error) {
	jsBody := string(body)
	iStart := strings.Index(jsBody, "(")
	iEnd := strings.LastIndex(jsBody, ")")

	if iStart == -1 || iEnd == -1 {
		return "", errors.New("Something is wrong with payload.")
	}

	return jsBody[iStart+1 : iEnd], nil
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
		Timeout:   globalTimeout,
		Transport: transport,
	}, nil
}

func (platform Platform) doRequest(lookupUrl string, stats chan<- Stat, errorsChannel chan *error) {
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

func resolveAndOpenGraph(url string) (stat Stat, urls []string, err error) {
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
	request.Header.Set("User-Agent", "Googlebot-News")
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
	stat, e = platforms[len(platforms)-1].parseWith(response)
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

func aggregateAndCombine(results map[string]interface{}, errors []error) map[string]interface{} {
	total := 0
	for _, p := range results {
		c := p.(map[string]interface{})["count"]
		if reflect.ValueOf(c).Kind() == reflect.Int {
			total += c.(int)
		} else if reflect.ValueOf(c).Kind() == reflect.Float64 {
			n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 64))
			total += n
		} else if reflect.ValueOf(c).Kind() == reflect.Float32 {
			n, _ := strconv.Atoi(strconv.FormatFloat(c.(float64), 'f', 0, 32))
			total += n
		} else {
			// logger.Fatal("Can't cast...")
		}
	}

	results["meta"] = map[string]interface{}{"total": total}

	errorsStrings := []string{}
	for _, error := range errors {
		errorsStrings = append(errorsStrings, error.Error())
	}

	results["errors"] = errorsStrings
	return results
}

var proxy = ""
var logger *log.Logger
var errorsLogger *log.Logger
var globalTimeout = time.Duration(4 * time.Second)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logger = log.New(ioutil.Discard, "socol ", log.Ldate|log.Ltime|log.Lshortfile)
		errorsLogger = log.New(ioutil.Discard, "socol ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "socol ", log.Ldate|log.Ltime|log.Lshortfile)
		errorsLogger = log.New(os.Stderr, "socol ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

func New(lookupUrl string, selectedPlatforms []string, privateProxy string) map[string]interface{} {
	proxy = privateProxy

	if selectedPlatforms == nil ||
		(len(selectedPlatforms) == 1 && selectedPlatforms[0] == "") {
		selectedPlatforms = []string{}
	}

	selectedPlatforms = append(selectedPlatforms, "origin")
	errors, stats, taskCount := make(chan *error), make(chan Stat), 0
	aggregated := map[string]interface{}{}
	errorsCollection := []error{}

	rStat, urls, rError := resolveAndOpenGraph(lookupUrl)
	if rError != nil {
		errorsLogger.Println(rError)
	} else {
		aggregated[rStat.name] = rStat.data
	}

	if len(urls) > 1 {
		logger.Println("Digging for", urls)
	}

	lookupUrl = urls[len(urls)-1]

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
			errorsCollection = append(errorsCollection, *error)
			taskCount--
		default:
			if taskCount <= 0 {
				return aggregateAndCombine(aggregated, errorsCollection)
			}
		}
	}
}
