.PHONY: all
all: lint test

.PHONY: lint
lint:
	go list -m -f '{{.Dir}}/...' | xargs golangci-lint run

.PHONY: test
test:
	go list -m -f '{{.Dir}}/...' | xargs go test

