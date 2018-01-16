Helm Monitor
============

> Monitor a release, rollback to a previous version depending on the result of
a PromQL (Prometheus), Lucene or DSL query (ElasticSearch).

## Install

```bash
$ helm plugin install https://github.com/ContainerSolutions/helm-monitor
```

## Usage

A rollback happen only if the number of result from the query is greater than 0.

You can find a step-by-step example in the `./examples` directory.

### Prometheus

Monitor the **peeking-bunny** release against a Prometheus server, a rollback
is initiated if the 5xx error rate is over 0 for the last minute.

```bash
$ helm monitor peeking-bunny 'rate(http_requests_total{status=~"^5..$"}[1m]) > 0'
```

You can connect to a given Prometheus instance, by default it will connect to
*http://localhost:9091*.

```bash
$ helm monitor --prometheus http://prometheus.domain.com:9091 \
    peeking-bunny \
    'rate(http_requests_total{status=~"^5..$"}[1m]) > 0'
```

### ElasticSearch

Monitor the **peeking-bunny** release against an ElasticSearch server, a
rollback is initiated if the 5xx error rate is over 0 for the last minute.

```bash
$ helm monitor peeking-bunny ''
```

You can connect to a given ElasticSearch instance, by default it will connect to
*http://localhost:9200*.

```bash
$ helm monitor --prometheus http://prometheus.domain.com:9091 \
    peeking-bunny \
    'rate(http_requests_total{status=~"^5..$"}[1m]) > 0'
```

## Development

Clone the repo, then add a symlink to the Helm plugin directory:

```bash
$ ln -s $GOPATH/src/github.com/ContainerSolutions/helm-monitor ~/.helm/plugins/helm-monitor
```

Install dependencies using [dep](https://github.com/golang/dep):

```bash
$ dep ensure
```

Build:

```bash
$ go build -o helm-monitor ./cmd/...
```

Run:

```bash
$ helm monitor elasticsearch my-release ./examples/elasticsearch-query.json
```
