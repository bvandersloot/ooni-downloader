package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("ooni-downloader")
var outputDirectory string
var getParameters map[string]string

const URL = "https://measurements.ooni.torproject.org/api/v1/files?"
const Retries = 10

type Metadata struct {
	count       int     `json:"count"`
	currentPage int     `json:"current_page"`
	limit       int     `json:"limit"`
	nextURL     *string `json:"next_url"`
	offset      int     `json:"offset"`
	pages       int     `json:"pages"`
}

type Result struct {
	downloadURL   string    `json:"download_url"`
	index         int       `json:"index"`
	probeASN      string    `json:"probe_asn"`
	probeCC       string    `json:"probe_cc"`
	testStartTime time.Time `json:"test_start_time"`
}

type Response struct {
	metadata Metadata `json:"metadata"`
	results  []Result `json:"result"`
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%s [flags] http_param1:value1 ... http_paramN:valueN\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	var format = logging.MustStringFormatter("%{level:.4s} %{time} %{longfunc}>    %{message}")
	logging.SetFormatter(format)
	flag.StringVar(&outputDirectory, "output-directory", "./", "what folder should contain all of the pulled data-files")
	flag.Parse()
	getParameters = make(map[string]string)
	for _, arg := range flag.Args() {
		if strings.Count(arg, ":") != 1 {
			flag.Usage()
			log.Fatal("Must use exactly one colon in HTTP GET parameter pairs")
		}
		arr := strings.Split(arg, ":")
		getParameters[arr[0]] = arr[1]
	}
}

func limit() int {
	if val, ok := getParameters["limit"]; ok {
		ret, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal("Integer not provided for limit argument")
		}
		return ret
	}
	return 100
}

func getWithRetry(url string) (*http.Response, error) {
	lastStatus := ""
	for i := 0; i < Retries; i++ {
		r, err := http.Get(url)
		if err != nil || r.StatusCode == 200 {
			return r, err
		}
		lastStatus = r.Status
		if i != Retries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil, errors.New(fmt.Sprintf("Too many failures when attempting to connect to %s. Current status: %s", url, lastStatus))
}

func producer(results chan Result) {
	currentURL := URL
	parameters := []string{}
	for k, v := range getParameters {
		parameters = append(parameters, fmt.Sprintf("%s=%s", k, v))
	}
	currentURL = currentURL + strings.Join(parameters, "&")
	for currentURL != "" {
		resp, err := getWithRetry(currentURL)
		if err != nil {
			log.Fatalf("Faulure when connecting to OONI: %s", err.Error())
		}
		dec := json.NewDecoder(resp.Body)
		var parsed Response
		if err := dec.Decode(&parsed); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Response did not comply to expected format. %s", err.Error())
		}
		if parsed.metadata.nextURL == nil {
			currentURL = ""
		} else {
			currentURL = *parsed.metadata.nextURL
		}
		for _, result := range parsed.results {
			results <- result
		}
		resp.Body.Close()
	}
}

func consumer(results chan Result, wg *sync.WaitGroup) {
	for result := range results {
		resp, err := getWithRetry(result.downloadURL)
		if err != nil {
			log.Fatalf("Faulure when connecting to data: %s", err.Error())
		}
		f, err := os.Create(filepath.Join(outputDirectory, fmt.Sprintf("%d", result.index)))
		if err != nil {
			log.Fatalf("Faulure making the output file: %s", err.Error())
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatalf("Faulure downloading to the output file: %s", err.Error())
		}
		f.Close()
		resp.Body.Close()
	}
	wg.Done()
}

func main() {
	os.MkdirAll(outputDirectory, os.ModePerm)
	communication := make(chan Result, limit())
	var wg sync.WaitGroup
	for i := 0; i < limit(); i++ {
		wg.Add(1)
		go consumer(communication, &wg)
	}
	producer(communication)
	close(communication)
	wg.Wait()
}
