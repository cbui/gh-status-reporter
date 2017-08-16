package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
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
	OrgRepo     string
	SHA         string
	Context     string
	Description string
	TargetUrl   string
	Username    string
	Auth        string
}

func validateRequiredFlags(flags Flags) error {
	if flags.OrgRepo == "" {
		return errors.New("Error: No Github organization/repository provided")
	}

	if flags.SHA == "" {
		return errors.New("Error: No SHA provided")
	}

	if flags.Context == "" {
		return errors.New("Error: No Github commit status context provided")
	}

	if flags.Auth == "" {
		return errors.New("Error: No auth token or password provided")
	}

	return nil
}

func setGithubCommitStatus(url string, flags Flags, state string) error {
	params := &CommitStatusParams{
		State:       state,
		TargetUrl:   flags.TargetUrl,
		Description: flags.Description,
		Context:     flags.Context,
	}

	requestBody, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("Error converting %q to json %s.", params, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	req.SetBasicAuth(flags.Username, flags.Auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error executing request to Github: %s", err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %q %s", resp.Body, err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Error creating commit status on Github.\n%s", responseBody)
	}

	return nil
}

func exitIfError(err error) {
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	orgRepo := flag.String("r", "", "Required: Github repository in the form of organization/repository, e.g google/cadvisor")
	sha := flag.String("s", "", "Required: Github commit status SHA")
	context := flag.String("c", "", "Required: Github commit status context")
	description := flag.String("d", "", "Optional: Github commit status description")
	targetUrl := flag.String("t", "", "Optional: Github commit status target_url")
	username := flag.String("u", "", "Optional: Github username for basic auth")
	auth := flag.String("a", "", "Required: Github password or token for basic auth")

	flag.Parse()

	flags := &Flags{
		OrgRepo:     *orgRepo,
		SHA:         *sha,
		Context:     *context,
		Description: *description,
		TargetUrl:   *targetUrl,
		Username:    *username,
		Auth:        *auth,
	}

	err := validateRequiredFlags(*flags)
	exitIfError(err)

	var cmd string
	var args []string

	if flag.NArg() > 0 {
		cmd = flag.Args()[0]
		args = flag.Args()[1:]
	} else {
		fmt.Printf("Error: no command given")
		os.Exit(1)
	}

	url := "https://api.github.com/repos/" + *orgRepo + "/statuses/" + *sha

	subprocess := exec.Command(cmd, args...)
	subprocess.Stdin, subprocess.Stdout, subprocess.Stderr = os.Stdin, os.Stdout, os.Stderr

	err = setGithubCommitStatus(url, *flags, "pending")
	exitIfError(err)

	err = subprocess.Run()

	if err == nil {
		err = setGithubCommitStatus(url, *flags, "success")
		exitIfError(err)
		os.Exit(0)
	}

	if err.Error() != "0" {
		err = setGithubCommitStatus(url, *flags, "failure")
		exitIfError(err)
		os.Exit(1)
	}

	if err != nil {
		err = setGithubCommitStatus(url, *flags, "error")
		exitIfError(err)
		fmt.Printf("Error: executing command %s with args %q: %s\n", cmd, args, err)
		os.Exit(1)
	}
}
