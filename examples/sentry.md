Rollback based on Sentry events
===============================

> Use helm-monitor to rollback a release based on Sentry events

## Prepare

Make sure to follow the steps described in [README.md](README.md) in order to
start Minikube with Tiller installed and pre-build the application.

Install Sentry, increase the timeout to let Tiller execute the chart hooks:

```bash
$ helm install --version 0.4.1 --name sentry stable/sentry --timeout 3200
$ minikube service sentry-sentry
```

Open the Sentry UI and configure a new project called "my-project".

Install the example application with the DSN:
```bash
$ helm upgrade -i my-app \
    --set image.tag=1.0.0 \
    --set env.SENTRY_DSN=<DSN> \
    --set env.SENTRY_RELEASE=1.0.0 \
    ./app/charts
```

### Upgrade and monitor

```bash
# get the Sentry endpoint
$ sentry=$(minikube service sentry-sentry --url)

# release version 2
$ helm upgrade -i my-app \
    --set image.tag=2.0.0 \
    --set env.SENTRY_DSN=<DSN> \
    --set env.SENTRY_RELEASE=2.0.0 \
    ./app/charts

# monitor
$ helm monitor sentry my-app \
    --api-key <SENTRY_API_KEY> \
    --organization sentry \
    --project my-project \
    --sentry $sentry \
    --release 2.0.0 \
    'Error triggered'
```

In a new terminal, simulate internal server failure:
```bash
$ curl $(minikube service my-app --url)/internal-error
```
