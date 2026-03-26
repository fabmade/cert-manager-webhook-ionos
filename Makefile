OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

IMAGE_NAME := "fabmade/cert-manager-webhook-ionos"
IMAGE_TAG := "1.3.0"

OUT := $(shell pwd)/_out
TEST := $(shell pwd)/_test

K8S_VERSION := 1.28.3
ENVTEST_BINDIR := $(TEST)/kubebuilder/bin/k8s/$(K8S_VERSION)-$(OS)-$(ARCH)
TEST_ASSET_KUBE_APISERVER := "$(ENVTEST_BINDIR)/kube-apiserver"
TEST_ASSET_ETCD := "$(ENVTEST_BINDIR)/etcd"
TEST_ASSET_KUBECTL := "$(ENVTEST_BINDIR)/kubectl"

$(shell mkdir -p "$(OUT)")

test: _test/kubebuilder
	TEST_ASSET_ETCD="$(TEST_ASSET_ETCD)" TEST_ASSET_KUBE_APISERVER="$(TEST_ASSET_KUBE_APISERVER)" TEST_ASSET_KUBECTL="$(TEST_ASSET_KUBECTL)" \
	go test -v .

_test/kubebuilder:
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	setup-envtest use $(K8S_VERSION) --bin-dir $(TEST)/kubebuilder/bin -p path

clean: clean-kubebuilder

clean-kubebuilder:
	rm -Rf $(TEST)/kubebuilder

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
	    --name example-webhook \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag=$(IMAGE_TAG) \
        deploy/example-webhook > "$(OUT)/rendered-manifest.yaml"

# Commands
docker: docker-build docker-push

docker-build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

docker-push:
	docker push ${IMAGE_NAME}:${IMAGE_TAG}
