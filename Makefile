
build:
	go build -o search-logger main.go

clean:
	go clean
	rm -f search-logger coverage.out

test:
	go test ./... -v

run:
	go run main.go