test:
	@go test -v ./...

tidy:  ## Get the dependencies
	@go mod tidy