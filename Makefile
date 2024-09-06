# Go parameters
GOCMD:=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

server-run:
	$(GORUN) runner/server/main.go

client-run:
	$(GORUN) runner/client/main.go

worker-run:
	$(GORUN) runner/worker/main.go $(ARGS)

server-build:
	docker build . -f deploy/server.Dockerfile -t go-streaming-server

client-build:
	docker build . -f deploy/client.Dockerfile -t go-streaming-client

worker-build:
	docker build . -f deploy/worker.Dockerfile -t go-streaming-worker

test:
	$(GOCMD) clean -testcache
	$(GOTEST) ./...

testc:
	$(GOCMD) clean -testcache
	$(GOTEST) -cover ./...

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/**/*.proto
