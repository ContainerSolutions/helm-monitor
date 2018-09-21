Helm monitor example
====================

> This example demonstrate how to use helm-monitor to rollback a Helm release
based on events.

## Prepare

Prepare you environment (Minikube, Tiller and build the required images):

```bash
# initialise Tiller with RBAC
$ kubectl create serviceaccount tiller -n kube-system
$ kubectl create clusterrolebinding tiller --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
$ helm init --wait

# build the application for Minikube
$ make build

# release version 1
$ helm upgrade -i my-app --set image.tag=1.0.0 ./app/charts

# access the application
$ minikube service my-app
```

## Follow the monitoring procedure from the example below

- [rollback based on Prometheus](prometheus.md)
- [rollback based on ElasticSearch](elasticsearch.md)
- [rollback based on Sentry](sentry.md)
