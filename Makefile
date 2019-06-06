.PHONY: test coverage test-race-detector help godoc lint

GO_PACKAGE_NAME:=gitlab.skypicker.com/go/sentryhook
COVERAGE_FILE:=$(TMP)/test-coverage.txt

#? test: run tests
test:
	go test $(shell go list ./... | grep -v /.cache/ ) -v -coverprofile ${COVERAGE_FILE}

#? test-race-detector: run tests with race detector enabled
test-race-detector:
	go test $(shell go list ./... | grep -v /.cache/ ) -race

#? coverage: run tests with coverage report
coverage: test
	go tool cover -func=${COVERAGE_FILE}

#? godoc: run a local GoDoc server
godoc:
	@echo Starting local GoDoc server on port 8888
	@echo Open http://localhost:8888/pkg/$(GO_PACKAGE_NAME) in your browser
	@# This is a temporary fix because GoDoc has issues with Go modules,
	@# so the solution at the moment is to fake GOROOT and GOPATH
	@mkdir -p /tmp/tmpgoroot/doc
	@rm -rf /tmp/tmpgopath/src/$(GO_PACKAGE_NAME)
	@mkdir -p /tmp/tmpgopath/src/$(GO_PACKAGE_NAME)
	@tar -c --exclude='.git' --exclude='tmp' . | tar -x -C /tmp/tmpgopath/src/$(GO_PACKAGE_NAME)
	@GOROOT=/tmp/tmpgoroot/ GOPATH=/tmp/tmpgopath/ godoc -http="localhost:8888"

#? lint: run a meta linter
lint:
	@hash golangci-lint || (echo "Download golangci-lint from https://github.com/golangci/golangci-lint#install" && exit 1)
	golangci-lint run

#? help: display help
help: Makefile
	@printf "Available make targets:\n\n"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sed -e 's/^/ /'