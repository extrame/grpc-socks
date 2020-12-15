version=$(shell git rev-parse --short HEAD)
buildAt=$(shell date "+%Y-%m-%d %H:%M:%S %Z")


build:
	rm -rf exec_bin
	mkdir exec_bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(version) -X 'main.buildAt=$(buildAt)'" -o ./exec_bin/server-linux ./server
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(version) -X 'main.buildAt=$(buildAt)'" -o ./exec_bin/client-linux ./client
idl:
	rm -rf pb/*.pb.go
	protoc -I=. pb/*.proto --go_out=plugins=grpc:.

dev:
	rsync ./exec_bin/client-linux root@101.201.199.194:
	rsync ./exec_bin/server-linux root@60.205.171.80: