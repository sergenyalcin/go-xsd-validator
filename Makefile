lint:
	golangci-lint run -v

# Run unit tests
unit-test:
	go test -v ./...

# Run tests with race detection
unit-test-race:
	go test -race -v ./...

# Run tests and generate a coverage report
unit-test-cover:
	go test -cover -v ./...
