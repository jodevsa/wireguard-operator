# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.0.1
SHELL = /bin/bash


# setup

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
KIND ?= $(LOCALBIN)/kind
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.8.0
KIND_VERSION ?= v0.19.0


# images
AGENT_IMAGE ?= "ghcr.io/jodevsa/wireguard-operator/agent:main"
MANAGER_IMAGE ?= "ghcr.io/jodevsa/wireguard-operator/manager:main"
SIDECAR_IMAGE ?= "ghcr.io/jodevsa/wireguard-operator/sidecar:main"

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# example.com/manager-bundle:$VERSION and example.com/manager-catalog:$VERSION.
IMAGE_TAG_BASE ?= ghcr.io/jodevsa/wireguard-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-operator-bundle:main

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.22

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

##@ Build
build: build-agent build-manager

build-agent: generate fmt vet ## Build manager binary.
	go build -o bin/agent ./cmd/agent/main.go

build-manager: generate fmt vet ## Build manager binary.
	go build -o bin/manager ./cmd/manager/main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/manager/main.go

docker-build-agent:  ## Build docker image with the manager.
	docker build -t ${AGENT_IMAGE} . -f ./images/agent/Dockerfile

docker-build-manager:  ## Build docker image with the manager.
	docker build -t ${MANAGER_IMAGE} . -f ./images/manager/Dockerfile

docker-build-sidecar:  ## Build docker image with the sidecar.
	docker build -t ${SIDECAR_IMAGE} . -f ./images/sidecar/Dockerfile

docker-build-all:
	$(MAKE) docker-build-agent
	$(MAKE) docker-build-manager
	${MAKE} docker-build-sidecar

docker-load-kind:
	kind load docker-image ${AGENT_IMAGE} ${SIDECAR_IMAGE} ${MANAGER_IMAGE}

run-e2e:
	SIDECAR_IMAGE=${SIDECAR_IMAGE} AGENT_IMAGE=${AGENT_IMAGE} $(MAKE) update-agent-and-sidecar-image
	MANAGER_IMAGE=${MANAGER_IMAGE} $(MAKE) update-manager-image
	$(KUSTOMIZE) build config/default > release_it.yaml
	git checkout ./config/default/manager_auth_proxy_patch.yaml
	git checkout ./config/manager/kustomization.yaml
	KUBE_CONFIG=$(HOME)/.kube/config KIND_BIN=${KIND} WIREGUARD_OPERATOR_RELEASE_PATH="../../release_it.yaml" SIDECAR_IMAGE=${SIDECAR_IMAGE} AGENT_IMAGE=${AGENT_IMAGE} MANAGER_IMAGE=${MANAGER_IMAGE} go test ./internal/it/ -v -count=1

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -


update-agent-and-sidecar-image: kustomize
	## TODO: Simplify later
	SIDECAR_IMAGE=$(SIDECAR_IMAGE) AGENT_IMAGE=$(AGENT_IMAGE) envsubst < ./config/default/manager_auth_proxy_patch.yaml.template > ./config/default/manager_auth_proxy_patch.yaml

update-manager-image: kustomize
	$(info MANAGER_IMAGE: "$(MANAGER_IMAGE)")
	cd config/manager && $(KUSTOMIZE) edit set image controller=${MANAGER_IMAGE}

generate-release-file: kustomize update-agent-image update-manager-image
	$(KUSTOMIZE) build config/default > release.yaml
	git checkout ./config/default/manager_auth_proxy_patch.yaml
	git checkout ./config/manager/kustomization.yaml

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

kind: $(KIND) ## Download kind locally if necessary.
$(KIND): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind/cmd/kind@$(KIND_VERSION)

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest