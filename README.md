# Template for new GO applications

This is basic skeleton to start new projects with.

Features:

* Easy configurable by environment variables and consul k/v storage
* Simple API with health check endpoints (echo)
* Prometheus metrics (with api endpoint to collect them)
* Logstash integration with logrus hook
* Sentry integration with logrus hook
* Pprof profiling can be enabled on demand
* Built and packaged into docker image

# Usage

## Start new project

Run `init.sh {YOUR_PROJECT_NAME} {YOUR_PROJECT_GOPATH}` to initialize new
project in `{YOUR_PROJECT_GOPATH}` directory with new name. New package path
should be relative to your GOPATH.
Please use single word names containing only lowercase letters and
dashes/underscores for project name.

Example: `./init.sh myapi github.com/account/myapi`

## Prerequisites

It is recommended to use Consul with this template.
Application requires `REGISTRY_DSN` environment variable to be provided.  
`REGISTRY_DSN` is consul API endpoint.
