package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
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

type Flags struct {
	OrgRepo     *string
	SHA         *string
	Context     *string
	Description *string
	TargetUrl   *string
	Username    *string
	Password    *string
}

func commitStatusURL(orgRepo string, sha string) string {
	return "https://api.github.com/repos/" + orgRepo + "/statuses/" + sha
}

func validateRequiredFlags(flags Flags) {
	errors := false

	if *flags.Username == "" || *flags.Password == "" {
		fmt.Println("Error: No username and password provided")
		errors = true
	}

	if *flags.OrgRepo == "" {
		fmt.Println("Error: No Github organization/repository provided")
		errors = true
	}

	if *flags.Context == "" {
		fmt.Println("Error: No SHA provided")
		errors = true
	}

	if *flags.Context == "" {
		fmt.Println("Error: No Github commit status context provided")
		errors = true
	}

	if errors {
		flag.Usage()
		os.Exit(1)
	}
}

func setGithubCommitStatus(url string, flags Flags, state string) {
	params := &CommitStatusParams{
		State:       state,
		TargetUrl:   *flags.TargetUrl,
		Description: *flags.Description,
		Context:     *flags.Context,
	}

	requestBody, err := json.Marshal(params)
	if err != nil {
		log.Fatalf("Error converting %q to json %s.", params, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	req.SetBasicAuth(*flags.Username, *flags.Password)

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
			url, *flags.Username, *flags.Password, requestBody, responseBody)
	}
}

func main() {
	flags := &Flags{
		OrgRepo:     flag.String("r", "", "Required: Github repository in the form of organization/repository, e.g google/cadvisor"),
		SHA:         flag.String("s", "", "Required: Github commit status SHA"),
		Context:     flag.String("c", "", "Required: Github commit status context"),
		Description: flag.String("d", "", "Optional: Github commit status description"),
		TargetUrl:   flag.String("t", "", "Optional: Github commit status target_url"),
		Username:    flag.String("u", "", "Required: Github username"),
		Password:    flag.String("p", "", "Required: Github password or token for basic auth"),
	}

	flag.Parse()

	validateRequiredFlags(*flags)

	var cmd string
	var args []string

	if flag.NArg() > 0 {
		cmd = flag.Args()[0]
		args = flag.Args()[1:]
	} else {
		log.Fatalf("Error: no command given")
	}

	url := commitStatusURL(*flags.OrgRepo, *flags.SHA)

	subprocess := exec.Command(cmd, args...)
	subprocess.Stdin, subprocess.Stdout, subprocess.Stderr = os.Stdin, os.Stdout, os.Stderr

	setGithubCommitStatus(url, *flags, "pending")

	err := subprocess.Run()

	if err == nil {
		setGithubCommitStatus(url, *flags, "success")
		os.Exit(0)
	}

	if err.Error() != "0" {
		setGithubCommitStatus(url, *flags, "failure")
		os.Exit(1)
	}

	if err != nil {
		setGithubCommitStatus(url, *flags, "error")
		log.Fatalf("Error: executing command %s with args %q: %s", cmd, args, err)
	}
}
