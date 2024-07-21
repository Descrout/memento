run:
	go run *.go

build:
	go build -ldflags "-s -w" -o main *.go

sync:
	git checkout autocomplete
	git pull origin autocomplete
	go mod tidy
	go build -ldflags "-s -w" -o main *.go
	sudo systemctl restart memento