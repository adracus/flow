.PHONY: fmt
fmt:
	gofumpt -w *.go

.PHONY: test
test:
	go test ./...

.PHONY: verify
verify:
	test fmt

