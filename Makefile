BINARY := prompt-evolver
MODULE := github.com/timholm/prompt-evolver

.PHONY: build test clean install lint

build:
	go build -o $(BINARY) .

test:
	go test -v -race ./...

clean:
	rm -f $(BINARY)

install: build
	mv $(BINARY) $(GOPATH)/bin/ 2>/dev/null || mv $(BINARY) ~/go/bin/

lint:
	go vet ./...
