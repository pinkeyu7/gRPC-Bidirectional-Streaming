# Go parameters
GOCMD:=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

run:
	$(GORUN) main.go

test:
	$(GOCMD) clean -testcache
	$(GOTEST) ./...

testc:
	$(GOCMD) clean -testcache
	$(GOTEST) -cover ./...

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/**/*.proto