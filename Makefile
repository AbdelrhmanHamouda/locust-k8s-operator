# VERSION is the operator release the OLM bundle describes. The bundle deploys
# the operator image with this tag, so it should be a published tag of
# lotest/locust-k8s-operator. bundle/ and bundle.Dockerfile are generated, not
# committed: the release workflow runs `make bundle VERSION=<tag>` on the tag
# it releases and attaches the result to the GitHub release. Locally it
# defaults to the newest release tag reachable from HEAD (0.0.1 if the clone
# has no tags); set it on the command line to build another version.
VERSION ?= $(shell git describe --tags --abbrev=0 --match '[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo 0.0.1)

# CHANNELS / DEFAULT_CHANNEL are the OLM channels written into
# bundle/metadata/annotations.yaml. DEFAULT_CHANNEL defaults to the first
# entry of CHANNELS, e.g. `make bundle CHANNELS=candidate,stable` makes
# candidate the default.
comma := ,
CHANNELS ?= stable
DEFAULT_CHANNEL ?= $(firstword $(subst $(comma), ,$(CHANNELS)))
BUNDLE_METADATA_OPTS ?= --channels=$(CHANNELS) --default-channel=$(DEFAULT_CHANNEL)

# IMAGE_TAG_BASE is the operator image repository. The bundle and catalog
# images are derived from it (<base>-bundle, <base>-catalog).
IMAGE_TAG_BASE ?= docker.io/lotest/locust-k8s-operator

# BUNDLE_OPERATOR_IMG is the operator image the generated CSV deploys. It
# defaults to the release image for VERSION. Passing IMG on the command line
# (`make bundle IMG=quay.io/me/locust-op:dev`, to test an unreleased build
# through OLM) uses that image instead.
ifeq ($(origin IMG),command line)
BUNDLE_OPERATOR_IMG ?= $(IMG)
else
BUNDLE_OPERATOR_IMG ?= $(IMAGE_TAG_BASE):$(VERSION)
endif

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# BUNDLE_K8S_VERSION is the Kubernetes version `make bundle-validate` checks
# the bundle against for removed APIs. It follows k8s.io/api in go.mod, the
# same minor version envtest runs (ENVTEST_K8S_VERSION, defined below).
BUNDLE_K8S_VERSION ?= $(ENVTEST_K8S_VERSION)

# OPENSHIFT_VERSIONS is the com.redhat.openshift.versions value `make bundle`
# adds to the bundle metadata. The OpenShift community-operators pipeline
# rejects a bundle with a minKubeVersion and no such value. "v4.16" means 4.16
# and later, which matches the CSV's minKubeVersion of 1.29.
OPENSHIFT_VERSIONS ?= v4.16

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif

# operator-sdk release downloaded into bin/ by `make operator-sdk`. Keep the
# scorecard-test image tags in config/scorecard/patches/ on the same version.
OPERATOR_SDK_VERSION ?= v1.42.3
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

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

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: test
test: manifests generate fmt vet setup-envtest generate-test-crds ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v -E '/(e2e|test/utils)$$') -coverprofile cover.out.tmp
	@grep -v -E '(zz_generated|test/utils)' cover.out.tmp > cover.out
	@rm -f cover.out.tmp

.PHONY: generate-test-crds
generate-test-crds: kustomize manifests ## Generate v1-only CRD for integration tests (no conversion webhook needed).
	cp config/crd/bases/locust.io_locusttests.yaml config/crd/test/base.yaml
	$(KUSTOMIZE) build config/crd/test > config/crd/test/locust.io_locusttests.yaml
	rm config/crd/test/base.yaml

# TODO(user): To use a different vendor for e2e tests, modify the setup under 'tests/e2e'.
# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# CertManager is installed by default; skip with:
# - CERT_MANAGER_INSTALL_SKIP=true
KIND_CLUSTER ?= locust-k8s-operator-test-e2e

.PHONY: setup-test-e2e
setup-test-e2e: ## Set up a Kind cluster for e2e tests if it does not exist
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "Kind is not installed. Please install Kind manually."; \
		exit 1; \
	}
	@case "$$($(KIND) get clusters)" in \
		*"$(KIND_CLUSTER)"*) \
			echo "Kind cluster '$(KIND_CLUSTER)' already exists. Skipping creation." ;; \
		*) \
			echo "Creating Kind cluster '$(KIND_CLUSTER)'..."; \
			$(KIND) create cluster --name $(KIND_CLUSTER) ;; \
	esac

.PHONY: test-e2e
test-e2e: setup-test-e2e manifests generate fmt vet ## Run the e2e tests. Expected an isolated environment using Kind.
	KIND_CLUSTER=$(KIND_CLUSTER) go test ./test/e2e/ -v -ginkgo.v
	$(MAKE) cleanup-test-e2e

.PHONY: cleanup-test-e2e
cleanup-test-e2e: ## Tear down the Kind cluster used for e2e tests
	@$(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

##@ CI

.PHONY: ci
ci: lint test ## Run all CI checks locally

.PHONY: ci-coverage
ci-coverage: test ## Generate coverage report for CI
	@echo "Coverage report: cover.out"
	@go tool cover -func=cover.out | tail -1

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager ./cmd

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd --enable-webhooks=false

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name locust-k8s-operator-builder
	$(CONTAINER_TOOL) buildx use locust-k8s-operator-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm locust-k8s-operator-builder
	rm Dockerfile.cross

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller (webhooks DISABLED — safe default, no cert-manager required).
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: deploy-with-webhook
deploy-with-webhook: manifests kustomize ## Deploy controller WITH admission webhook (requires cert-manager pre-installed).
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default-webhook | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: undeploy-with-webhook
undeploy-with-webhook: kustomize ## Undeploy controller deployed via deploy-with-webhook.
	$(KUSTOMIZE) build config/default-webhook | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.8.2
CONTROLLER_TOOLS_VERSION ?= v0.22.0
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.14.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

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

# operator-sdk ships as a release binary rather than a `go install`-able
# module, so it gets its own download helper. Same layout as go-install-tool:
# a versioned binary plus an unversioned symlink in $(LOCALBIN). The download
# is checked against the release's checksums.txt; update the hashes below
# together with OPERATOR_SDK_VERSION.
#
# Set OPERATOR_SDK to use a binary you already have. The target then leaves it
# alone and downloads nothing.
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
OPERATOR_SDK_SHA256_darwin_amd64 := 7cb0f24bb63b6383a117291ee4c808953c5dd789d5877da98051aa68b41f40ac
OPERATOR_SDK_SHA256_darwin_arm64 := 098ae8b9dbe7dfd557e8e7ed0f1996736922dd4b984621df2aa033f225cae161
OPERATOR_SDK_SHA256_linux_amd64 := 887a3bb0d63ccc4ca47a522d0c8ffac56d9d5246f6a2bd886b4ed23eb2e2672f
OPERATOR_SDK_SHA256_linux_arm64 := 6db93cd821b429f0bb514cea4bbb5553827d273fc8aa211f13e14798599d31cd

.PHONY: operator-sdk
operator-sdk: ## Download operator-sdk into bin/ if necessary (skipped when OPERATOR_SDK is set).
ifeq ($(OPERATOR_SDK),$(LOCALBIN)/operator-sdk)
operator-sdk: $(LOCALBIN)
	@[ -f "$(OPERATOR_SDK)-$(OPERATOR_SDK_VERSION)" ] || { \
	set -e ;\
	OS=$$(go env GOOS) ; ARCH=$$(go env GOARCH) ;\
	case "$$OS-$$ARCH" in \
		darwin-amd64) want=$(OPERATOR_SDK_SHA256_darwin_amd64) ;; \
		darwin-arm64) want=$(OPERATOR_SDK_SHA256_darwin_arm64) ;; \
		linux-amd64) want=$(OPERATOR_SDK_SHA256_linux_amd64) ;; \
		linux-arm64) want=$(OPERATOR_SDK_SHA256_linux_arm64) ;; \
		*) echo "No pinned operator-sdk checksum for $$OS/$$ARCH; set OPERATOR_SDK to your own binary." >&2 ; exit 1 ;; \
	esac ;\
	echo "Downloading operator-sdk $(OPERATOR_SDK_VERSION)" ;\
	tmp="$(OPERATOR_SDK)-$(OPERATOR_SDK_VERSION).tmp" ;\
	curl -sSfL -o "$$tmp" \
		"https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH}" ;\
	got=$$( { command -v sha256sum >/dev/null && sha256sum "$$tmp" || shasum -a 256 "$$tmp" ; } | cut -d' ' -f1 ) ;\
	if [ "$$got" != "$$want" ]; then \
		rm -f "$$tmp" ;\
		echo "operator-sdk checksum mismatch: got $$got, want $$want" >&2 ; exit 1 ;\
	fi ;\
	chmod +x "$$tmp" ;\
	mv "$$tmp" "$(OPERATOR_SDK)-$(OPERATOR_SDK_VERSION)" ;\
	} ;\
	ln -sf "$(OPERATOR_SDK)-$(OPERATOR_SDK_VERSION)" "$(OPERATOR_SDK)"
endif

##@ OLM bundle

# bundle/ and bundle.Dockerfile are generated output (ignored by git), so the
# target starts from scratch: operator-sdk only ever adds files, and a resource
# dropped from config/ would otherwise keep shipping in bundle/manifests.
#
# The operator image comes from BUNDLE_OPERATOR_IMG through config/olm's
# `images:` entry, which is set for the build and restored afterwards so the
# tracked kustomization doesn't change. config/olm resets the manager image
# first, so whatever `make deploy IMG=...` left in config/manager can't leak
# in, and config/manifests copies the final image into the CSV's
# containerImage annotation.
.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate and validate the OLM bundle (bundle/, bundle.Dockerfile) for VERSION. Operator image: BUNDLE_OPERATOR_IMG, or IMG if given.
	rm -rf bundle/manifests bundle/metadata bundle/tests bundle.Dockerfile
	$(OPERATOR_SDK) generate kustomize manifests -q --interactive=false
	set -e ;\
	cp config/olm/kustomization.yaml config/olm/kustomization.yaml.orig ;\
	trap 'mv config/olm/kustomization.yaml.orig config/olm/kustomization.yaml' EXIT ;\
	(cd config/olm && $(KUSTOMIZE) edit set image controller=$(BUNDLE_OPERATOR_IMG)) ;\
	manifests="$$($(KUSTOMIZE) build config/manifests)" ;\
	printf '%s\n' "$$manifests" | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	grep -q '^  com.redhat.openshift.versions:' bundle/metadata/annotations.yaml || \
		printf '\n  # OpenShift versions the bundle supports.\n  com.redhat.openshift.versions: "%s"\n' '$(OPENSHIFT_VERSIONS)' >> bundle/metadata/annotations.yaml
	grep -q '^LABEL com.redhat.openshift.versions=' bundle.Dockerfile || \
		printf '\n# OpenShift versions the bundle supports.\nLABEL com.redhat.openshift.versions="%s"\n' '$(OPENSHIFT_VERSIONS)' >> bundle.Dockerfile
	$(MAKE) bundle-validate

.PHONY: bundle-validate
bundle-validate: operator-sdk ## Validate bundle/ with the checks the OperatorHub and OpenShift community pipelines run.
	$(OPERATOR_SDK) bundle validate ./bundle
	$(OPERATOR_SDK) bundle validate ./bundle --select-optional suite=operatorframework --optional-values=k8s-version=$(BUNDLE_K8S_VERSION)

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_TOOL) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = $(LOCALBIN)/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.73.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool $(CONTAINER_TOOL) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
