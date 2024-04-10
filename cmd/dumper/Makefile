# Define variables
hash = $(shell git rev-parse --short HEAD)
registry = ghcr.io/jacobbrewer1/dumpster
DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

summary:
	@echo "Building summary"
	make linux
	chmod +x bin/linux
PR-approval:
	@echo "Running PR CI"
	go build ./...
	go vet ./...
	go test ./...
ci: docker
	# Push the image to the registry
	docker push $(registry):$(hash) && docker push $(registry):latest
linux: clean
	@echo "Building for linux"
	GOOS=linux GOARCH=amd64 go build -o bin/linux -ldflags '-X main.Commit=$(hash) -X main.Date=$(DATE)'
docker: linux
	@echo "Building docker image"
	# Build the docker image
	docker build -t $(registry):$(hash) -t $(registry):latest .
clean:
	@echo "Cleaning up"
	# Remove the bin directory
	rm -rf bin
