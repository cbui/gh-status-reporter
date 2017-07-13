package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func defaultFlags() *Flags {
	return &Flags{
		OrgRepo:     "christopher-bui/gh-status-reporter",
		SHA:         "deadbeef",
		Context:     "ci",
		Description: "unit test",
		TargetUrl:   "",
		Username:    "octocat",
		Auth:        "token",
	}
}

func TestValidateRequiredFlags(t *testing.T) {
	flags := defaultFlags()

	err := validateRequiredFlags(*flags)
	if err != nil {
		t.Errorf("Got error even though all required flags are present.\n%s", err.Error())
	}

	requiredFields := []string{"OrgRepo", "SHA", "Context", "Auth"}
	for _, field := range requiredFields {
		flags = defaultFlags()
		reflect.ValueOf(flags).Elem().FieldByName(field).SetString("")

		err = validateRequiredFlags(*flags)
		if err == nil {
			t.Errorf("Should have gotten error with missing %s\n", field)
		}
	}
}

func TestSetGithubCommitStatusHappyPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedMethod := "POST"
		if r.Method != expectedMethod {
			t.Errorf("Expected request method to be %q, got %q", expectedMethod, r.Method)
		}

		basicAuth := r.Header.Get("Authorization")
		expectedAuth := "Basic b2N0b2NhdDp0b2tlbg=="
		if basicAuth != expectedAuth {
			t.Errorf("Expected 'Authorization' header value to be: %q, got %q", expectedAuth, basicAuth)
		}

		body, _ := ioutil.ReadAll(r.Body)
		expectedBody := "{\"state\":\"pending\",\"target_url\":\"\",\"description\":\"unit test\",\"context\":\"ci\"}"
		if string(body[:]) != expectedBody {
			t.Errorf("Expected request body to be: %q, got %q", expectedBody, body)
		}

		fmt.Fprintln(w, "Ok")
	}))
	defer ts.Close()

	setGithubCommitStatus(ts.URL, *defaultFlags(), "pending")
}

func TestSetGithubCommitStatusCommitStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404 - Not Found")
	}))
	defer ts.Close()

	err := setGithubCommitStatus(ts.URL, *defaultFlags(), "pending")
	expectedError := "Error creating commit status on Github.\n404 - Not Found\n"
	if err.Error() != expectedError {
		t.Errorf("Expected to get an error: %q, got %q", expectedError, err.Error())
	}
}
