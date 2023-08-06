build:
	go build -o bin/simple-pass .

run:
	go run

generate:
	go generate ./...

tidy:
	go mod tidy

test:
	go test -v -cover ./...

benchmark:
	go test -v -cover ./...


integration-test: 
	./integration-tests/*.sh

full-build-and-test: build test integration-test
