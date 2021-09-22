# `/internal`

Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself. See the Go 1.4 [`release notes`](https://golang.org/doc/go1.4#internalpackages) for more details. Note that you are not limited to the top level `internal` directory. You can have more than one `internal` directory at any level of your project tree.

Examples:

* https://github.com/hashicorp/terraform/tree/master/internal
* https://github.com/influxdata/influxdb/tree/master/internal
* https://github.com/perkeep/perkeep/tree/master/internal
* https://github.com/jaegertracing/jaeger/tree/master/internal
* https://github.com/moby/moby/tree/master/internal
* https://github.com/satellity/satellity/tree/master/internal

## `/internal/pkg`

Examples:

* https://github.com/hashicorp/waypoint/tree/main/internal/pkg