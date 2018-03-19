.PHONY: installtiller
installtiller:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.9/plugin/pkg/auth/authorizer/rbac/bootstrappolicy/testdata/cluster-roles.yaml --validate=false
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.9/plugin/pkg/auth/authorizer/rbac/bootstrappolicy/testdata/cluster-role-bindings.yaml --validate=false
	kubectl create serviceaccount tiller --namespace kube-system || true
	kubectl create clusterrolebinding tiller \
		--clusterrole=cluster-admin \
		--serviceaccount=kube-system:tiller || true
	helm init --upgrade --service-account tiller

.PHONY: prepare
prepare:
	eval $$(minikube docker-env) && \
	docker build --build-arg "VERSION=1.0.0" -t app:1.0.0 app && \
	docker build --build-arg "VERSION=2.0.0" -t app:2.0.0 app

.PHONY: deployv1
deployv1:
	helm upgrade --install my-release ./app/charts --set image.tag=1.0.0

.PHONY: deployv2
deployv2:
	helm upgrade my-release ./app/charts --set image.tag=2.0.0

.PHONY: installprometheus
installprometheus:
	helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
	helm upgrade --install prometheus-operator coreos/prometheus-operator
	kubectl apply -f ./prometheus.yaml

.PHONY: cleanup
	kubectl delete -f ./prometheus.yaml
	helm del --purge prometheus-operator my-release
