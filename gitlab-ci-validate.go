package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

var DEFAULT_HOST = "https://gitlab.com"

type config struct {
	Host      string `json:"host"`
	Token     string `json:"token"`
	ProjectId string `json:"project_id"`
}

type lintRequest struct {
	Content string `json:"content"`
}

type lintResponse struct {
	Valid  bool     `json:"valid"`
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
func ValidateFile(targetUrl *url.URL, path string) (Validation, []error) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("fail to read file %s: %s\n", path, err)
		return HARD_FAIL, []error{err}
	}

	if !strings.HasSuffix(path, ".yml") {
		return HARD_FAIL, []error{fmt.Errorf("file name does not end with .yml - only .gitlab-ci.yaml is allowed by GitLab")}
	}

	var body interface{}
	if err := yaml.Unmarshal([]byte(content), &body); err != nil {
		return HARD_FAIL, []error{err}
	}

	lr := lintRequest{Content: string(content)}
	lrJson, err := json.Marshal(lr)
	if err != nil {
		return HARD_FAIL, []error{err}
	}

	request, err := http.NewRequest("POST", targetUrl.String(), bytes.NewBuffer(lrJson))
	if err != nil {
		return SOFT_FAIL, []error{err}
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return SOFT_FAIL, []error{err}
	}
	defer response.Body.Close()
	if response.StatusCode == 401 || response.StatusCode == 403 {
		fmt.Printf("HTTP %d recieved from %s, authentication is required. See usage on how to provide an identity if you have not already, otherwise double check your basic auth or token.\n", response.StatusCode, targetUrl.Host)
		responseBytes, err := io.ReadAll(response.Body)
		if err == nil {
			fmt.Printf("message from server: %s\n", responseBytes)
		}
		os.Exit(1)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return SOFT_FAIL, []error{err}
	}

	if response.StatusCode != 200 {
		return SOFT_FAIL, []error{fmt.Errorf("Non-200 status from %s for %s: %d: %s", targetUrl.Host, targetUrl.Path, response.StatusCode, responseBytes)}
	}

	var summary lintResponse
	if err := json.Unmarshal(responseBytes, &summary); err != nil {
		return SOFT_FAIL, []error{err}
	}
	if !summary.Valid {
		errs := make([]error, len(summary.Errors))
		for i, err := range summary.Errors {
			errs[i] = errors.New(err)
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
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	flags.Usage = func() {
		fmt.Printf("Usage: %s [-host=string] [-token=string] [-project-id=string] FILE...\n", os.Args[0])
		flags.PrintDefaults()
	}

	var c config
	flags.StringVar(&c.Host, "host", getEnv("GITLAB_HOST", DEFAULT_HOST), "GitLab instance used to validate the config files")
	flags.StringVar(&c.Token, "token", getEnv("GITLAB_TOKEN", ""), "GitLab API access token")
	flags.StringVar(&c.ProjectId, "project-id", getEnv("GITLAB_PROJECT_ID", ""), "GitLab project ID")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	targetUrl, err := url.Parse(c.Host)
	if err != nil {
		fmt.Printf("host '%s' is not valid URL: %s\n", c.Host, err)
		os.Exit(1)
	}
	if targetUrl.Scheme == "" {
		targetUrl.Scheme = "https"
		// this is because the baseUrl.Host is not set when the scheme is no present
		targetUrl, err = url.Parse(targetUrl.String())
		if err != nil {
			fmt.Printf("host '%s' is not a valid URL: %s\n", targetUrl.String(), err)
		}
	}
	if c.ProjectId == "" {
		fmt.Printf("project-id is required\n")
		os.Exit(1)
	} else {
		targetUrl.Path = fmt.Sprintf("/api/v4/projects/%s/ci/lint", c.ProjectId)
	}
	if c.Token == "" && targetUrl.User == nil {
		fmt.Printf("token is required\n")
		os.Exit(1)
	} else {
		query := targetUrl.Query()
		query.Add("private_token", c.Token)
		targetUrl.RawQuery = query.Encode()
	}

	l := log.New(os.Stderr, "", 0)
	if flags.NArg() < 1 {
		flags.Usage()
		os.Exit(1)
	}

	var result Validation
	for _, source := range flags.Args() {
		validation, errs := ValidateFile(targetUrl, source)
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
