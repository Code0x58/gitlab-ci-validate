package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type lintResponse struct {
	Status string   `json:"status"`
	Errors []string `json:"errors"`
}

type Validation int

const (
	// file passed remote validation
	PASS Validation = iota
	// file couldn't be remotely validated
	SOFT_FAIL
	// file failed local or remote validation
	HARD_FAIL
)

var version string
var userAgent string

func init() {
	userAgent = fmt.Sprintf("gitlab-ci-validate/%s go/%s %s/%s", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Validate the given file
func ValidateFile(hostUrl url.URL, path string) (Validation, []error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return HARD_FAIL, []error{err}
	}

	if !strings.HasSuffix(path, ".yml") {
		return HARD_FAIL, []error{fmt.Errorf("file name does not end with .yml - only .gitlab-ci.yaml is allowed by GitLab")}
	}

	data, err := yaml.YAMLToJSON(content)
	if err != nil {
		return HARD_FAIL, []error{err}
	}

	values := url.Values{"content": {string(data)}}
	hostUrl.Path = "/api/v4/ci/lint"
	request, err := http.NewRequest("POST", hostUrl.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return SOFT_FAIL, []error{err}
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return SOFT_FAIL, []error{err}
	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		fmt.Printf("HTTP 401 recieved from %s, authentication is required. See usage on how to provide an identity if you have not already, otherwise double check your basic auth or token.\n", hostUrl.Host)
		os.Exit(1)
	}
	if response.StatusCode != 200 {
		return SOFT_FAIL, []error{fmt.Errorf("Non-200 status from %s for %s: %d", hostUrl.Host, hostUrl.Path, response.StatusCode)}
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return SOFT_FAIL, []error{err}
	}

	var summary lintResponse
	json.Unmarshal(responseBytes, &summary)
	if summary.Status != "valid" {
		errs := make([]error, len(summary.Errors))
		for i, err := range summary.Errors {
			errs[i] = fmt.Errorf(err)
		}
		return HARD_FAIL, errs
	}
	return PASS, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-host=string] [-token=string] file ...\n", os.Args[0])
		flag.PrintDefaults()
	}

	token := flag.String("token", getEnv("GITLAB_TOKEN", ""), "GitLab API access token")

	host := flag.String("host", getEnv("GITLAB_HOST", "https://gitlab.com"), "GitLab instance used to validate the config files")
	flag.Parse()

	baseUrl, err := url.Parse(*host)
	if err != nil {
		fmt.Printf("host is not valid URL: %s\n", *host)
		os.Exit(1)
	}
	if baseUrl.Scheme == "" {
		baseUrl.Scheme = "https"
		// this is because the baseUrl.Host is not set when the scheme is no present
		baseUrl, err = url.Parse(baseUrl.String())
		if err != nil {
			fmt.Printf("host is not a valid URL: %s\n", *host)
		}
	}
	if *token != "" {
		params := url.Values{"private_token": {*token}}
		baseUrl.RawQuery = params.Encode()
	}

	l := log.New(os.Stderr, "", 0)
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	var result Validation
	for _, source := range flag.Args() {
		validation, errs := ValidateFile(*baseUrl, source)
		if validation > result {
			result = validation
		}
		if errs == nil {
			l.Printf("PASS: %s\n", source)
		} else {
			var status string
			if validation == SOFT_FAIL {
				status = "SOFT"
			} else {
				status = "HARD"
			}
			l.Printf("%s FAIL: %s\n", status, source)
			for _, err := range errs {
				l.Printf(" - %s\n", err)
			}
		}
	}
	if result == HARD_FAIL {
		os.Exit(1)
	} else if result == SOFT_FAIL {
		os.Exit(2)
	} else {
		os.Exit(0)
	}
}
