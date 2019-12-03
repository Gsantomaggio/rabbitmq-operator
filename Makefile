
# Image URL to use all building/pushing image targets
IMG_BASE ?= rabbitmq-operator
IMG ?= ${IMG_BASE}:2

GIT_HUB_IMG ?= docker.pkg.github.com/gsantomaggio/rabbitmq-operator/${IMG}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push docker.pkg.github.com/gsantomaggio/rabbitmq-operator/${IMG}


docker-push-on-dockerhub:
	docker build . -t ${IMG}
	docker build . -t ${IMG_BASE}:latest
	
	docker tag ${IMG} gsantomaggio/${IMG}
	docker push gsantomaggio/${IMG}
	
	docker tag ${IMG_BASE}:latest  gsantomaggio/${IMG_BASE}:latest
	docker push gsantomaggio/${IMG_BASE}:latest
	
	cd config/manager && kustomize edit set image controller=${IMG} 
	kustomize build config/default  > deploy/rabbitmq-operator_tag.yaml
	
	cd config/manager && kustomize edit set image controller=${IMG_BASE}:latest 
	kustomize build config/default  > deploy/rabbitmq-operator_latest.yaml

docker-push-on-github:
	docker build . -t ${GIT_HUB_IMG}
	docker push ${GIT_HUB_IMG}
	cd config/manager && kustomize edit set image controller=${GIT_HUB_IMG} 
	kustomize build config/default  > deploy/rabbitmq-operator_gh.yaml


# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
