.PHONY: all
all: requirements verify

.PHONY: requirements
requirements:
	@./hack/install-requirements.sh

.PHONY: fmt
fmt:
	@gofumpt -w *.go

.PHONY: test
test:
	@go test ./...

.PHONY: check
check:
	@go vet ./...
	@./hack/check-format.sh *.go

.PHONY: verify
verify: check test fmt

