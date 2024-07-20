run:
	go run *.go

build:
	go build -ldflags "-s -w" -o main *.go