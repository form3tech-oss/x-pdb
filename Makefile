##########
# Docker #
##########

# Image names
IMG                       ?= ghcr.io/form3tech-oss/x-pdb:latest
WEBHOOK_IMG               ?= ghcr.io/form3tech-oss/x-pdb-webhook:latest
TEST_APP_IMG              ?= x-pdb-test:latest
TEST_DISRUPTION_PROBE_IMG ?= x-pdb-test-disruption-probe:latest

# Docker file paths
DOCKERFILE_PATH                       ?= Dockerfile
WEBHOOK_DOCKERFILE_PATH               ?= Dockerfile.webhook
TEST_APP_DOCKERFILE_PATH              ?= Dockerfile.testapp
TEST_DISRUPTION_PROBE_DOCKERFILE_PATH ?= Dockerfile.testdisruptionprobe

########
# Kind #
########

CONTEXT            ?= kind-$(KIND_CLUSTER_NAME)
CLUSTER            ?= 1
KIND_IMAGE         ?= kindest/node:v1.31.2
KIND_CLUSTER_NAME  ?= x-pdb-$(CLUSTER)

#########
# Proto #
#########

PROTO_FILES := $(wildcard proto/**/**/*.proto)
PROTO_GO_OUT_DIR := pkg

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
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
	$(CONTROLLER_GEN) crd webhook paths="./..." output:crd:artifacts:config=charts/x-pdb/crds

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: ## Run tests.
	go test -v -race $(shell go list ./... | grep -v tests) -coverprofile cover.out

##@ E2E Tests

.PHONY: test-app-docker-build
test-app-docker-build: ## Generates Test App docker image.
	$(MAKE) docker-build IMG=$(TEST_APP_IMG) DOCKERFILE_PATH=$(TEST_APP_DOCKERFILE_PATH)

.PHONY: test-app-load-image
test-app-load-image: ## Loads Test App docker image into a KinD cluster.
	kind load docker-image $(TEST_APP_IMG) --name $(KIND_CLUSTER_NAME)

.PHONY: test-disruption-probe-docker-build
test-disruption-probe-docker-build: ## Generates Test disruption probe docker image.
	$(MAKE) docker-build IMG=$(TEST_DISRUPTION_PROBE_IMG) DOCKERFILE_PATH=$(TEST_DISRUPTION_PROBE_DOCKERFILE_PATH)

.PHONY: test-disruption-probe-load-image
test-disruption-probe-load-image: ## Loads Test disruption probe docker image into a KinD cluster.
	kind load docker-image $(TEST_DISRUPTION_PROBE_IMG) --name $(KIND_CLUSTER_NAME)

.PHONY: deploy-e2e
deploy-e2e: test-app-docker-build test-disruption-probe-docker-build webhook-docker-build docker-build ## Deploys x-pdb and loads test images into all the testing KinD clusters.
	@echo "building and deploying x-pdb and e2e test apps"
	for number in 1 2 3; do \
		$(MAKE) install CLUSTER=$$number; \
		$(MAKE) test-app-load-image CLUSTER=$$number; \
		$(MAKE) test-disruption-probe-load-image CLUSTER=$$number; \
	done

.PHONY: e2e
e2e: ## Runs the E2E tests on top of the KinD clusters.
	go test -v -race -timeout 30m ./tests/...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: helm-lint
helm-lint: # Lints the x-pdb helm chartk
	helm lint charts/x-pdb

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/controller/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -f $(DOCKERFILE_PATH) -t $(IMG) .

.PHONY: webhook-docker-build
webhook-docker-build: ## Generates Test disruption probe docker image.
	$(MAKE) docker-build IMG=$(WEBHOOK_IMG) DOCKERFILE_PATH=$(WEBHOOK_DOCKERFILE_PATH)

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' $(DOCKERFILE_PATH) > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name x-pdb-builder
	$(CONTAINER_TOOL) buildx use x-pdb-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm x-pdb-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: multi-cluster
multi-cluster: gen-certs ## Creates all the testing KinD clusters.
	for number in 1 2 3; do \
		$(MAKE) kind-cluster install-metallb install-cert-manager CLUSTER=$$number; \
	done

.PHONY: destroy-multi-cluster
destroy-multi-cluster: ## Destroys all the testing KinD clusters.
	kind get clusters | grep x-pdb | xargs -I {} kind delete cluster -n {}

.PHONY: deploy
deploy: docker-build webhook-docker-build ## Deploys x-pdb on all testing KinD clusters.
	@echo "building and deploying x-pdb"
	for number in 1 2 3; do \
		$(MAKE) install CLUSTER=$$number; \
	done

.PHONY: kind-cluster
kind-cluster: ## Creates a KinD cluster.
	echo "CREATING CLUSTER context=$(CONTEXT) cluster=$(CLUSTER)"
	kind create cluster --config=./hack/env/kind-${CLUSTER}.yaml --image $(KIND_IMAGE) --name $(KIND_CLUSTER_NAME)

.PHONY: install-metallb
install-metallb: ## Installs metallb on a KinD cluster.
	./hack/install-metallb.sh $(CONTEXT) $(CLUSTER)

.PHONY: install-cert-manager
install-cert-manager: ## Installs cert-manager and a cluster issuer in a KinD cluster.
	kubectl apply --wait=true -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml --context $(CONTEXT)
	kubectl wait deployment -n cert-manager cert-manager-webhook --for condition=Available=True --timeout=90s --context $(CONTEXT)
	kubectl create secret tls ca-key-pair -n cert-manager --cert=hack/certs/ca.crt --key=hack/certs/ca.key --dry-run=client -o yaml | kubectl apply -f - --context $(CONTEXT)
	kubectl apply -f hack/env/cluster-issuer.yaml --wait=true --context $(CONTEXT)

.PHONY: kind-load
kind-load: ## Loads an image into a KinD cluster.
	kind load docker-image ${IMG} --name $(KIND_CLUSTER_NAME)

.PHONY: kind-load-webhook
kind-load-webhook: ## Loads Test App docker image into a KinD cluster.
	kind load docker-image $(WEBHOOK_IMG) --name $(KIND_CLUSTER_NAME)

.PHONY: gen-certs
gen-certs: ## Generates all the TLS certificates for x-pdb
	./hack/gen-certs.sh

.PHONY: install
install: kind-load kind-load-webhook ## Installs x-pdb into a cluster
	./hack/install-xpdb.sh $(CONTEXT) $(CLUSTER)

##@ Proto

.PHONY: proto-generate
proto-generate: ## Generates the go packages from the proto contracts.
	@buf generate

.PHONY: proto-lint
proto-lint: ## Lints the proto contracts
	@buf lint

.PHONY: proto-fmt
proto-fmt: ## Formats the proto contracts
	@buf format -w

.PHONY: proto-breaking
proto-breaking: ## Verifies if there are beaking changes in the proto contracts
	@buf breaking --against 'https://github.com/form3tech-oss/x-pdb.git'

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.4
GOLANGCI_LINT_VERSION ?= v1.61.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: protoc-gen-go
protoc-gen-go: $(PROTOC_GEN_GO) ## Download controller-gen locally if necessary.
$(PROTOC_GEN_GO): $(LOCALBIN)
	$(call go-install-tool,$(PROTOC_GEN_GO),google.golang.org/protobuf/cmd/protoc-gen-go,$(PROTOC_GEN_GO_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
