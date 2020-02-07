SHELL=/bin/bash -o pipefail
GO_PKG=github.com/vernemq/vmq-operator
FIRST_GOPATH:=$(firstword $(subst :, ,$(shell go env GOPATH)))
GOJSONTOYAML_BINARY:=$(FIRST_GOPATH)/bin/gojsontoyaml

all: $(GOJSONTOYAML_BINARY) generate


.PHONY: generate
generate: operator_sdk_gen jsonnet/vmq-operator/vernemq-crd.libsonnet examples docs

.PHONY: docs
docs:
	go run github.com/vernemq/vmq-operator/cmd/apidoc > docs/api.md

jsonnet/vmq-operator/%-crd.libsonnet: $(shell find deploy/crds/*_crd.yaml -type f) $(GOJSONTOYAML_BINARY)
	cat deploy/crds/vernemq_v1alpha1_vernemq_crd.yaml | gojsontoyaml -yamltojson > jsonnet/vmq-operator/vernemq-crd.libsonnet

.PHONY: examples
examples:
	cd examples; \
	./build.sh vernemq-grafana-prometheus.jsonnet vernemq_grafana_prometheus; \
	./build.sh only-vernemq.jsonnet only_vernemq

.PHONY: operator_sdk_gen
operator_sdk_gen: operator_sdk_gen_k8s operator_sdk_gen_openapi

.PHONY: operator_sdk_gen_k8s
operator_sdk_gen_k8s:
	operator-sdk generate k8s

.PHONY: operator_sdk_gen_openapi
operator_sdk_gen_openapi:
	operator-sdk generate openapi

$(GOJSONTOYAML_BINARY):
	@go get github.com/brancz/gojsontoyaml
