BIN := lilcaster

.PHONY: build test tidy clean

build:
	go build -o $(BIN) .

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -f $(BIN)