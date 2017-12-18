# gh-status-reporter

[![Build Status](https://travis-ci.org/Christopher-Bui/gh-status-reporter.svg?branch=master)](https://travis-ci.org/Christopher-Bui/gh-status-reporter)

gh-status-reporter allows you to execute a command and reports the
result to GitHub as a commit status.

This can be useful in if you're using Docker MultiStage Builds and
need a way to report to Github from within a Docker build.

# Usage

```
Usage of ./gh-status-reporter:
  -a string
    	Required: Github password or token for basic auth
  -c string
    	Required: Github commit status context
  -d string
    	Optional: Github commit status description
  -dev string
    	Optional: If true, then ignores required flags and executes command as-is; without any status reporting
  -r string
    	Required: Github repository in the form of organization/repository, e.g google/cadvisor
  -s string
    	Required: Github commit status SHA
  -t string
    	Optional: Github commit status target_url
  -u string
    	Optional: Github username for basic auth
```

Instead of passing in a value for every flag, you may choose to use environment
variables that correspond to each flag:

```
BUILD_ORG_REPO
BUILD_SHA
BUILD_CONTEXT
BUILD_DESCRIPTION
BUILD_TARGET_URL
BUILD_USER
BUILD_AUTH
BUILD_DEV
```

```
Example:

go run main.go -r christopher-bui/gh-status-reporter \
  -c "docker/ci/test" \
  -a $GH_TOKEN \
  -s $SHA \
  sleep 25
```

After running that, provided you gave a valid sha and auth token, you
will have a pending commit status on that SHA. Then when the command
exits after 25 seconds, it will turn success.
