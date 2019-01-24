package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type lintResponse struct {
	Status string   `json:"status"`
	Errors []string `json:"errors"`
}

func main() {
	failed := false

	for _, source := range os.Args[1:] {
		// TODO(Code0x58): implement human friendly CLI
		// TODO(Code0x58): return consistent and human friendly errors
		content, err := ioutil.ReadFile(source)
		if err != nil {
			failed = true
			panic(err)
		}

		data, err := yaml.YAMLToJSON(content)
		if err != nil {
			failed = true
			panic(err)
		}
		response, err := http.PostForm("https://gitlab.com/api/v4/ci/lint", url.Values{"content": {string(data)}})
		if err != nil {
			failed = true
			panic(err)
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			failed = true
			panic(fmt.Errorf("Non-200 status from GitLab: %d", response.StatusCode))
		}

		responseBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			failed = true
			panic(err)
		}

		var summary lintResponse
		json.Unmarshal(responseBytes, &summary)
		if summary.Status != "valid" {
			failed = true
			fmt.Printf("%s failed:\n", source)
			for _, err := range summary.Errors {
				fmt.Println(err)
			}
		}
	}
	if failed {
		os.Exit(1)
	}
}
