package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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

// Validate the given file
func ValidateFile(path string) (Validation, []error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return HARD_FAIL, []error{err}
	}

	data, err := yaml.YAMLToJSON(content)
	if err != nil {
		return HARD_FAIL, []error{err}
	}
	response, err := http.PostForm("https://gitlab.com/api/v4/ci/lint", url.Values{"content": {string(data)}})
	if err != nil {
		return SOFT_FAIL, []error{err}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return SOFT_FAIL, []error{fmt.Errorf("Non-200 status from GitLab: %d", response.StatusCode)}
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

func main() {
	// TODO(Code0x58): return 1 if any are invalid, return 2 if only failures were with connecting to GitLab
	l := log.New(os.Stderr, "", 0)
	if len(os.Args) < 2 {
		l.Println("You must provide the paths to one or more GitLab CI config files.")
		os.Exit(1)
	}

	var result Validation
	for _, source := range os.Args[1:] {
		// TODO(Code0x58): implement human friendly CLI
		// TODO(Code0x58): return consistent and human friendly errors
		validation, errs := ValidateFile(source)
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
