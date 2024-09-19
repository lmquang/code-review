# Makefile for code-review

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=$$GOBIN/code-review

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/code-review

# Run the application
run:
	$(GOCMD) run ./cmd/code-review

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Build and run
build_and_run: build
	./$(BINARY_NAME)

# Run code review
review: build
	@if [ -z "$$OPENAI_API_KEY" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is not set"; \
		echo "Please set it with: export OPENAI_API_KEY='your-api-key'"; \
		exit 1; \
	fi
	./$(BINARY_NAME)

.PHONY: build run clean test build_and_run review