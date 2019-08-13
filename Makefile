# Image URL to use all building/pushing image targets
IMG ?= ray:0.1.0
all: ray

# Build manager binary
ray-controller: generate fmt
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/ray-controller github.com/ray-operator/cmd/ray-controller/
	
# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...
# Generate code
generate:
	go generate ./cmd/... ./pkg/...
# Build the docker image
docker-build:
	docker build . -t ${IMG}
# Push the docker image
docker-push:
	docker push ${IMG}
