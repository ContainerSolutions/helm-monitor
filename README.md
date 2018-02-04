Helm Monitor plugin
===================

> Monitor a release, rollback to a previous version depending on the result of
a PromQL (Prometheus), Lucene or DSL query (ElasticSearch).

<a href="https://asciinema.org/a/SKViqDmByVLgv14F9hrlYX9qs" target="_blank"><img src="https://asciinema.org/a/SKViqDmByVLgv14F9hrlYX9qs.png" style="width:100%" /></a>

![Helm monitor failure](helm-monitor-failure.png)

## Install

```bash
$ helm plugin install https://github.com/ContainerSolutions/helm-monitor
```

## Usage

A rollback happen only if the number of result from the query is greater than 0.

You can find a step-by-step example in the `./examples` directory.

### Prometheus

Monitor the **peeking-bunny** release against a Prometheus server, a rollback
is initiated if the 5xx error rate is over 0 as measured over the last 5
minutes.

```bash
$ helm monitor prometheus peeking-bunny 'rate(http_requests_total{code=~"^5.*$"}[5m]) > 0'
```

You can connect to a given Prometheus instance, by default it will connect to
*http://localhost:9090*.

```bash
$ helm monitor prometheus --prometheus=http://prometheus:9090 \
    peeking-bunny \
    'rate(http_requests_total{code=~"^5.*$"}[5m]) > 0'
```

### ElasticSearch

Monitor the **peeking-bunny** release against an ElasticSearch server, a
rollback is initiated if the 5xx error rate is over 0 for the last minute.

Using a Lucene query:

```bash
$ helm monitor elasticsearch peeking-bunny 'status:500 AND kubernetes.labels.app:app AND version:2.0.0'
```

Using a query DSL file:

```bash
$ helm monitor elasticsearch peeking-bunny ./query.json
```

You can connect to a given ElasticSearch instance, by default it will connect to
*http://localhost:9200*.

```bash
$ helm monitor elasticsearch --elasticsearch=http://elasticsearch:9200 \
    peeking-bunny \
    'status:500 AND kubernetes.labels.app:app AND version:2.0.0'
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

## Todo

- Investigate tool for failure detection compared to previous time window
- Add flag for [fancy graph](https://github.com/gizak/termui)
