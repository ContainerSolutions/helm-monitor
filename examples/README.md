Helm monitor example
====================

> In this example, we run the Prometheus operator and a GoLang application in
Minikube, upgrade then monitor for HTTP failure. If the amount of 5xx reach a
certain limit, then the application get automatically rolled back to its
previous state.


## Prepare

Build and release the GoLang application to Minikube:

```
$ eval $(minikube docker-env)
$ docker build -t app app
$ helm upgrade -i my-release ./app/charts
```

Access the application:

```
$ minikube service my-release-app
```

## Prometheus

### Setup

Install Prometheus using Prometheus operator:

```
$ helm install --name prometheus-operator coreos/prometheus-operator
$ kubectl apply -f ./prometheus.yaml
```

Access Prometheus:

```
$ minikube service prometheus
```

### Upgrade and monitor

```
$ kubectl port-forward prometheus-prometheus-0 9090
$ helm monitor my-release 'rate(http_requests_total{code=~"^4.*$"}[1m]) > 0'
```

Simulate failure:

```
$ export APP=$(minikube service my-release-app --url)
$ while sleep 0.1; do curl $APP/fail; done
```


## ElasticSearch

### Setup

Minikube support the EFK stack via addons, to enable it:

```
$ minikube addons enable efk
```

Access Kibana:

```
$ minikube service kibana-logging
```

### Upgrade and monitor

```
$ kubectl port-forward -n kube-system $(kubectl get po -n kube-system -l k8s-app=elasticsearch-logging -o jsonpath="{.items[0].metadata.name}") 9200
```

Test using via query DSL:
```
$ helm monitor elasticsearch my-release ./elasticsearch-query.json
```

Or via Lucene query
```
$ helm monitor elasticsearch my-release "status:500 AND kubernetes.labels.app:app"
```

Simulate failure:
```
$ export APP=$(minikube service my-release-app --url)
$ while sleep 0.1; do curl $APP/fail; done
```
