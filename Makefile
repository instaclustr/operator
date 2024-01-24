# Image URL to use all building/pushing image targets
IMG ?= operator:latest

IMG_TAG ?= latest
OPERATOR_NAMESPACE ?= instaclustr-operator
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.2
KUBEVIRT_TAG?=v1.57.0
KUBEVIRT_VERSION?=v1.0.0
ARCH?=linux-amd64

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: default

.PHONY: default
default: cert-deploy docker-build docker-push deploy

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test-clusters
test-clusters:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./controllers/clusters -coverprofile cover.out

.PHONY: test-clusterresources
test-clusterresources:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./controllers/clusterresources -coverprofile cover.out

.PHONY: test-kafkamanagement
test-kafkamanagement:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./controllers/kafkamanagement -coverprofile cover.out

.PHONY: test-users
test-users:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./controllers/tests

.PHONY: test-webhooks
test-webhooks:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./apis/clusters/v1beta1 -coverprofile cover.out

.PHONY: test
 test: manifests generate fmt vet docker-build-server-stub run-server-stub envtest test-clusters test-clusterresources test-kafkamanagement test-users stop-server-stub

.PHONY: goimports
goimports:
	goimports -v -w -local github.com/instaclustr/operator -l $(GO_FILES)

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	golangci-lint run --timeout 5m --skip-dirs=pkg/instaclustr/mock

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: manifests generate test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-build-server-stub
docker-build-server-stub: ## Build an Instaclustr server-stub image.
	docker build -f pkg/instaclustr/mock/server/Dockerfile --network=host -t instaclustr-server-stub .

.PHONY: run-server-stub
run-server-stub: ## Run an Instaclustr server-stub from your host.
	docker run --rm -d --name instaclustr-server-stub -p 8082:8082 instaclustr-server-stub

PHONY: stop-server-stub
stop-server-stub: ## Stop an Instaclustr server-stub from your host.
	docker stop instaclustr-server-stub

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd scripts && ./make_creds_secret.sh
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: helm-deploy
helm-deploy:
	helm repo add operator-chart https://instaclustr.github.io/operator-helm/
	helm repo up
	helm upgrade --install operator operator-chart/operator \
	--set image.tag=${IMG_TAG} --set USERNAME=${USER_NAME} --set APIKEY=${API_KEY} \
	--set HOSTNAME=${HOST_NAME} -n ${OPERATOR_NAMESPACE} --create-namespace --debug

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20231005234617-5771399a8ce5

.PHONY: cert-deploy
cert-deploy: ## Deploy cert-manager
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml

.PHONY: cert-undeploy
cert-undeploy: ## UnDeploy cert-manager
	kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml

.PHONY dev-build:
dev-build: docker-build kind-load deploy ## builds docker-image, loads it to kind cluster and deploys operator

.PHONY: kind-load
kind-load: ## loads given image to kind cluster
	kind load docker-image ${IMG}

.PHONY: deploy-kubevirt
deploy-kubevirt: ## Deploy kubevirt operator and custom resources
	kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml
	kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
	kubectl create -f https://github.com/kubevirt/containerized-data-importer/releases/download/${KUBEVIRT_TAG}/cdi-operator.yaml
	kubectl create -f https://github.com/kubevirt/containerized-data-importer/releases/download/${KUBEVIRT_TAG}/cdi-cr.yaml

.PHONY: undeploy-kubevirt
undeploy-kubevirt: ## Delete kubevirt operator and custom resources
	kubectl delete apiservices v1.subresources.kubevirt.io --ignore-not-found
	kubectl delete mutatingwebhookconfigurations virt-api-mutator --ignore-not-found
	kubectl delete validatingwebhookconfigurations virt-operator-validator --ignore-not-found
	kubectl delete validatingwebhookconfigurations virt-api-validator --ignore-not-found
	kubectl delete -n kubevirt kubevirt kubevirt --ignore-not-found
	kubectl delete -f https://github.com/kubevirt/containerized-data-importer/releases/download/${KUBEVIRT_TAG}/cdi-cr.yaml --ignore-not-found
	kubectl delete -f https://github.com/kubevirt/containerized-data-importer/releases/download/${KUBEVIRT_TAG}/cdi-operator.yaml --ignore-not-found
	kubectl delete -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml --ignore-not-found
	kubectl delete -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml --ignore-not-found

.PHONY: install-virtctl
install-virtctl: ## Install virtctl tool
	curl -L -o virtctl https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/virtctl-${KUBEVIRT_VERSION}-${ARCH}
	chmod +x virtctl
	sudo install virtctl /usr/local/bin
	rm virtctl
