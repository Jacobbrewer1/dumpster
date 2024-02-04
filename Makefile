CI-all: ci

PR-approval:
	@echo "Running PR CI"
	go build ./...
	go vet ./...
	go test ./...
ci: clean
	# For each subdirectory of the cmd directory, run make ci
	for d in cmd/*; do \
		(cd $$d && make ci); \
	done
	# Clean up
	make clean
clean:
	@echo "Cleaning up"
	# Loop through all subdirectories of the cmd directory and run make clean
	for d in cmd/*; do \
		(cd $$d && make clean); \
	done
