Rollback based on an Elasticsearch query
========================================

> Use helm-monitor to rollback a release based on an Elasticsearch query

## Prepare

Make sure to follow the steps described in [README.md](README.md) in order to
start Minikube with Tiller installed and pre-build the application.

Minikube support the EFK stack via addons, to enable it:

```bash
$ minikube addons enable efk
```

If Minikube was already running, you might need to restart it in order to have
the EFK stack up and running:

```bash
$ minikube stop
$ minikube start
```

Access Kibana:

```bash
$ minikube service kibana-logging -n kube-system
```

## Upgrade and monitor

```bash
# port forward elasticsearch port locally
$ kubectl port-forward -n kube-system $(kubectl get po -n kube-system -l k8s-app=elasticsearch-logging -o jsonpath="{.items[0].metadata.name}") 9200

# release version 2
$ helm upgrade -i my-app --set image.tag=2.0.0 ./app/charts

# monitor
$ helm monitor elasticsearch my-app ./elasticsearch-query.json

# or via Lucene query
$ helm monitor elasticsearch my-app "status:500 AND kubernetes.labels.app:app AND version:2.0.0"
```

Simulate internal server failure:

```bash
$ curl $(minikube service my-app --url)internal-error
```

## Cleanup

Delete my-app Helm releases:

```bash
$ helm del --purge my-app
```
