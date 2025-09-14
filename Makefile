module := go.fm
bin := go.fm
pkg := ./...

gofmt := gofmt
goimports := goimports
golangci_lint := golangci-lint

# default target
.PHONY: all
all: tidy fmt vet lint test build

# run the project
.PHONY: run
run:
	@echo ">> running $(bin)..."
	go run .

# build the binary
.PHONY: build
build:
	@echo ">> building $(bin)..."
	go build -o bin/$(bin) .

# install binary to gopath/bin
.PHONY: install
install:
	@echo ">> installing $(bin)..."
	go install .

# clean build artifacts
.PHONY: clean
clean:
	@echo ">> cleaning build artifacts..."
	rm -rf bin

# format code using gofmt and goimports
.PHONY: fmt
fmt:
	@echo ">> formatting source code..."
	$(gofmt) -s -w .
	@command -v $(goimports) >/dev/null 2>&1 && $(goimports) -w . || echo ">> goimports not installed"

# check for formatting issues without applying
.PHONY: fmt-check
fmt-check:
	@echo ">> checking formatting..."
	@$(gofmt) -l .
	@command -v $(goimports) >/dev/null 2>&1 && $(goimports) -l . || echo ">> goimports not installed"

# tidy up go.mod and go.sum
.PHONY: tidy
tidy:
	@echo ">> tidying modules..."
	go mod tidy

# run go vet for static analysis
.PHONY: vet
vet:
	@echo ">> running go vet..."
	go vet $(pkg)

# run linting (requires golangci-lint)
.PHONY: lint
lint:
	@echo ">> running linter..."
	@command -v $(golangci_lint) >/dev/null 2>&1 && $(golangci_lint) run ./... || echo ">> golangci-lint not installed"

# run tests with coverage
.PHONY: test
test:
	@echo ">> running tests..."
	go test -v -race -cover $(pkg)

# run tests and generate coverage report
.PHONY: cover
cover:
	@echo ">> generating coverage report..."
	go test -coverprofile=coverage.out $(pkg)
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo ">> open coverage.html in your browser to view detailed report"

# update all dependencies
.PHONY: update
update:
	@echo ">> updating dependencies..."
	go get -u ./...
	go mod tidy

# check for outdated dependencies
.PHONY: outdated
outdated:
	@echo ">> checking outdated dependencies..."
	@go list -u -m -json all | go run golang.org/x/exp/cmd/modoutdated

# run benchmarks
.PHONY: bench
bench:
	@echo ">> running benchmarks..."
	go test -bench=. -benchmem $(pkg)

# generate docs using go doc
.PHONY: docs
docs:
	@echo ">> generating documentation..."
	go doc -all $(module)
