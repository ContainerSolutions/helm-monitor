Helm monitor example
====================

> In this example, we run the Prometheus operator and a GoLang application in
Minikube, upgrade then monitor for HTTP failure. If the amount of 5xx reach a
certain limit, then the application get automatically rolled back to its
previous state.


## Prepare

```
# initialise Tiller
$ helm init --wait

# build the application for Minikube
$ make build

# release version 1
$ helm upgrade -i my-app --set image.tag=1.0.0 ./app/charts

# access the application
$ minikube service my-app
```

## Prometheus

### Setup

Install Prometheus:

```
$ helm install \
    --version 7.0.2 \
    --set server.service.type=Loadbalancer \
    --set server.global.scrape_interval=30s \
    --set alertmanager.enabled=false \
    --set kubeStateMetrics.enabled=false \
    --set nodeExporter.enabled=false \
    --set pushgateway.enabled=false \
    --name prometheus \
    stable/prometheus
```

Access Prometheus:

```
$ minikube service prometheus
```

### Upgrade and monitor

```
# get Prometheus endpoint
$ prometheus=$(minikube service prometheus-server --url)

# release version 2
$ helm upgrade -i my-app --set image.tag=2.0.0 ./app/charts

# monitor
$ helm monitor prometheus my-app --prometheus $prometheus 'rate(http_requests_total{code=~"^5.*$",version="2.0.0"}[5m]) > 0'
```

In a new terminal, simulate internal server failure:

```
$ app=$(minikube service my-app --url)
$ while sleep 1; do curl "$app"/internal-error; done
```

## ElasticSearch

### Setup

Minikube support the EFK stack via addons, to enable it:

```
$ minikube addons enable efk
```

If Minikube was already running, you might need to restart it in order to have
the EFK stack up and running:

```
$ minikube stop
$ minikube start
```

Access Kibana:

```
$ minikube service kibana-logging -n kube-system
```

### Upgrade and monitor

```
$ kubectl port-forward -n kube-system $(kubectl get po -n kube-system -l k8s-app=elasticsearch-logging -o jsonpath="{.items[0].metadata.name}") 9200
$ helm upgrade -i my-app --set image.tag=2.0.0 ./app/charts
```

Monitor using via query DSL:

```
$ helm monitor elasticsearch my-app ./elasticsearch-query.json
```

Or via Lucene query

```
$ helm monitor elasticsearch my-app "status:500 AND kubernetes.labels.app:app AND version:2.0.0"
```

Simulate internal server failure:

```
$ app=$(minikube service my-app --url)
$ while sleep 1; do curl "$app"/internal-error; done
```


## Cleanup

Delete Prometheus and my-app Helm releases:

```
$ make cleanup
```
