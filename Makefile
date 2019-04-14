.PHONY: fmt
fmt:
	gofumpt -w *.go

.PHONY: test
test:
	go test ./...

.PHONY: check
check:
	go vet ./...

.PHONY: verify
verify: check test fmt

