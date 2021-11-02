lint:
	@gofmt -w -s .
	@goimports -w .
	@go vet ./...
	@go mod tidy
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1 run

run:
	@go run main.go