package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type CommitStatusParams struct {
	State       string `json:"state"`
	TargetUrl   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

func commitStatusURL(orgRepo string, sha string) string {
	return "https://api.github.com/repos/" + orgRepo + "/statuses/" + sha
}

func setGithubCommitStatus(url string, username string, password string, params *CommitStatusParams) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		log.Fatalf("Error converting %q to json %s.", params, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	req.SetBasicAuth(username, password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error executing request to Github: %s", err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %q %s", resp.Body, err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Error creating commit status on Github.\nPOST %s\nAuth: '%s:%s'\nParams: %s\nResponse Body: %s",
			url, username, password, requestBody, responseBody)
	}
}

func main() {
	orgRepo := flag.String("r", "", "Github repository in the form of organization/repository, e.g google/cadvisor")
	sha := flag.String("s", "", "Github commit status SHA")
	context := flag.String("c", "", "Github commit status context")
	description := flag.String("d", "", "Github commit status description")
	targetUrl := flag.String("t", "", "Github commit status target_url")
	username := flag.String("u", "", "Github username")
	password := flag.String("p", "", "Github password or token for basic auth")

	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatalf("Error: no username or password provided")
	}

	var cmd string
	var args []string

	if flag.NArg() > 0 {
		cmd = flag.Args()[0]
		args = flag.Args()[1:]
	} else {
		log.Fatalf("Error: no command given")
	}

	url := commitStatusURL(*orgRepo, *sha)
	params := &CommitStatusParams{
		State:       "pending",
		TargetUrl:   *targetUrl,
		Description: *description,
		Context:     *context,
	}

	subprocess := exec.Command(cmd, args...)
	subprocess.Stdin, subprocess.Stdout, subprocess.Stderr = os.Stdin, os.Stdout, os.Stderr

	setGithubCommitStatus(url, *username, *password, params)

	err := subprocess.Run()

	if err == nil {
		params.State = "success"
		setGithubCommitStatus(url, *username, *password, params)
		os.Exit(0)
	}

	if err.Error() != "0" {
		params.State = "failure"
		setGithubCommitStatus(url, *username, *password, params)
		os.Exit(1)
	}

	if err != nil {
		params.State = "error"
		setGithubCommitStatus(url, *username, *password, params)
		log.Fatalf("Error: executing command %s with args %q: %s", cmd, args, err)
	}
}
