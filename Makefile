OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

IMAGE_NAME := "fabmade/cert-manager-webhook-ionos"
IMAGE_TAG := "1.0.3"

OUT := $(shell pwd)/_out
TEST := $(shell pwd)/_test

KUBEBUILDER_VERSION := 2.3.2
TEST_ASSET_KUBE_APISERVER := "$(TEST)/kubebuilder/bin/kube-apiserver"
TEST_ASSET_ETCD := "$(TEST)/kubebuilder/bin/etcd"
TEST_ASSET_KUBECTL := "$(TEST)/kubebuilder/bin/kubectl"

$(shell mkdir -p "$(OUT)")

test: _test/kubebuilder
	TEST_ASSET_ETCD="$(TEST_ASSET_ETCD)" TEST_ASSET_KUBE_APISERVER="$(TEST_ASSET_KUBE_APISERVER)" TEST_ASSET_KUBECTL="$(TEST_ASSET_KUBECTL)" \
	go test .

_test/kubebuilder:
	curl -fsSL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH).tar.gz -o kubebuilder-tools.tar.gz
	mkdir -p $(TEST)/kubebuilder
	tar -xvf kubebuilder-tools.tar.gz
	mv kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH)/bin $(TEST)/kubebuilder/
	rm kubebuilder-tools.tar.gz
	rm -R kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH)

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
