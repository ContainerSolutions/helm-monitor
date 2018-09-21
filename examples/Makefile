PROMETHEUS_VERSION=7.0.2

.PHONY: build
build:
	eval $$(minikube docker-env) && \
	docker build \
		--build-arg "LDFLAGS=-X main.version=1.0.0" \
		-t my-app:1.0.0 \
		app && \
	docker build \
		--build-arg "LDFLAGS=-X main.version=2.0.0" \
		-t my-app:2.0.0 \
		app

.PHONY: deployv1
deployv1:
	helm upgrade -i my-app --set image.tag=1.0.0 ./app/charts

.PHONY: deployv2
deployv2:
	helm upgrade -i my-app --set image.tag=2.0.0 ./app/charts

.PHONY: installprometheus
installprometheus:
	helm upgrade -i \
			--version $(PROMETHEUS_VERSION) \
			--set server.service.type=LoadBalancer \
			--set server.global.scrape_interval=30s \
			--set alertmanager.enabled=false \
			--set kubeStateMetrics.enabled=false \
			--set nodeExporter.enabled=false \
			--set pushgateway.enabled=false \
			prometheus \
			stable/prometheus

.PHONY: cleanup
	helm del --purge prometheus my-app
