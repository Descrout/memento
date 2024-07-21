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

ssh:
	chmod 400 ssh-key-2022-11-27.key
	ssh -i "ssh-key-2022-11-27.key" ubuntu@141.144.250.63

scp_upload:
	scp -i "ssh-key-2022-11-27.key" $(file) ubuntu@141.144.250.63:/home/ubuntu

scp_download:
	scp -i "ssh-key-2022-11-27.key" ubuntu@141.144.250.63:$(file) $(HOME)

.PHONY: ssh run build sync scp_upload scp_download