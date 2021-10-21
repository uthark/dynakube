
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

CRD_OPTIONS ?= "crd"

CONTROLLER_GEN_VERSION = 0.7.0
KUBEBUILDER_VERSION = 3.1.0
KUBERNETES_VERSION = 1.21.4

GOPROXY = https://proxy.golang.org,direct
GOPRIVATE = github.com/zendesk/*

OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)

SUCCESS_MSG = \033[0;32mSuccess!\033[0m

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Lint code
lint:
	golangci-lint run -v  --timeout=5m && echo "${SUCCESS_MSG}"
.PHONY: lint

# Run tests
test: go-testcov fmt vet
	$(GOTESTCOV) -mod=readonly ./... && echo "${SUCCESS_MSG}"
.PHONY: test

# Show code coverage
test/cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
.PHONY: test/cover

# Build manager binary
build:  fmt vet
	go build && echo "${SUCCESS_MSG}"
.PHONY=build

# Run go fmt
fmt:
	go fmt ./...
.PHONY: fmt

# Run go vet
vet:
	go vet ./...
.PHONY: vet

# Generate go code using go generate.
go-generate:
	go generate -x ./...
.PHONY: go-generate


# Build the docker image
docker-build: test
	docker build . -t ${IMG}
.PHONY: docker-build

# Push the docker image
docker-push:
	docker push ${IMG}
.PHONY: docker-push

# Download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION} ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
.PHONY: controller-gen

# Update API Docs.
docs:
	go install github.com/robertkrimen/godocdown/godocdown@latest
	godocdown ./api/v1alpha1 > API.md
.PHONY: docs

# Install Kube Builder.
install-kubebuilder:
	
	# download kubebuilder and extract it to tmp
	curl -Lo kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${OS}_${ARCH}

	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	mkdir -p /usr/local/kubebuilder/bin
	sudo mv kubebuilder /usr/local/kubebuilder/bin
	echo "Update your path: export PATH=${PATH}:/usr/local/kubebuilder/bin"
.PHONY: install-kubebuilder

# Install kubebuilder tools to support test runs.
install-kubebuilder-tools:
	curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz"
	mkdir -p /usr/local/kubebuilder
	tar -C /usr/local/kubebuilder --strip-components=1 -zvxf envtest-bins.tar.gz
.PHONY: install-kubebuilder-tools

# Install go-testcov binary.
go-testcov:
ifeq (, $(shell which go-testcov))
	@{ \
	set -e ;\
	go install github.com/grosser/go-testcov@v1.2.0;\
	}
GOTESTCOV=$(GOBIN)/go-testcov
else
GOTESTCOV=$(shell which go-testcov)
endif
.PHONY: go-testcov

# Show outdated dependencies
outdated:
	go install github.com/psampaz/go-mod-outdated@v0.8.0
	go list -u -m -json all | go-mod-outdated -direct
.PHONY: outdated

# Show this help
help:
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
.PHONY: help

