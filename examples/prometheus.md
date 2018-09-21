Rollback based on a Prometheus query
====================================

> Use helm-monitor to rollback a release based on a Prometheus query. In this
example we run a Prometheus instance and a GoLang application in Minikube,
upgrade then monitor for HTTP failure. If the amount of 5xx reach a certain
limit, then the application get automatically rolled back to its previous state.

## Prepare

Make sure to follow the steps described in [README.md](README.md) in order to
start Minikube with Tiller installed and pre-build the application.

Install Prometheus:

```bash
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

```bash
$ minikube service prometheus
```

### Upgrade and monitor

```bash
# get Prometheus endpoint
$ prometheus=$(minikube service prometheus-server --url)

# release version 2
$ helm upgrade -i my-app --set image.tag=2.0.0 ./app/charts

# monitor
$ helm monitor prometheus my-app --prometheus $prometheus 'rate(http_requests_total{code=~"^5.*$",version="2.0.0"}[5m]) > 0'
```

In a new terminal, simulate internal server failure:

```bash
$ app=$(minikube service my-app --url)
$ while sleep 1; do curl "$app"/internal-error; done
```

## Cleanup

Delete Prometheus and my-app Helm releases:

```bash
$ helm del --purge prometheus my-app
```
