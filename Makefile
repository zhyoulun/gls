fmt:
	go fmt ./...

vet:
	go vet ./...

ut:
	go test -gcflags=-l -race -covermode=atomic ./...

lint:
	golangci-lint run \
-E goimports \
-D deadcode -D unused -D structcheck -D varcheck ./...



tools:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1

