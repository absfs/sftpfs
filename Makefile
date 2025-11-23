.PHONY: test test-unit test-integration docker-up docker-down clean

# Run unit tests only
test-unit:
	go test -v -race -coverprofile=coverage.out

# Start Docker SFTP server
docker-up:
	docker-compose up -d
	@echo "Waiting for SFTP server to be ready..."
	@sleep 3

# Stop Docker SFTP server
docker-down:
	docker-compose down -v

# Run integration tests (requires Docker)
test-integration: docker-up
	go test -v -race -tags=integration -timeout 2m
	$(MAKE) docker-down

# Run all tests
test: test-unit test-integration

# Clean up
clean:
	rm -rf testdata/upload/*
	rm -f coverage.out
	$(MAKE) docker-down
