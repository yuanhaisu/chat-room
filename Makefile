RECV_SERVER=recv-server
SEND_SERVER=send-server
CLIENT1=client1

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
FLAG := "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse HEAD`"

run: $(RECV_SERVER)-$(ARCH)-$(OS) $(SEND_SERVER)-$(ARCH)-$(OS) $(CLIENT1)-$(ARCH)-$(OS)
	nohup ./$(SEND_SERVER)-$(ARCH)-$(OS) > $(SEND_SERVER).log 2>&1 &
	nohup ./$(RECV_SERVER)-$(ARCH)-$(OS) > $(RECV_SERVER).log 2>&1 &
	./$(CLIENT1)-$(ARCH)-$(OS)

build:
	@go clean
	@export GO111MODULE=on
	@echo "building for $(OS)/$(ARCH)"
	GOOS=$(OS)  GOARCH=$(ARCH) go build -ldflags $(FLAG) -x -o $(CLIENT1)-$(ARCH)-$(OS) chat-room/cmd/client
	GOOS=$(OS)  GOARCH=$(ARCH) go build -ldflags $(FLAG) -x -o $(RECV_SERVER)-$(ARCH)-$(OS) chat-room/cmd/recv_server
	GOOS=$(OS)  GOARCH=$(ARCH) go build -ldflags $(FLAG) -x -o $(SEND_SERVER)-$(ARCH)-$(OS) chat-room/cmd/send_server


prepare:
	yum install -y gcc
	curl -sSL https://get.daocloud.io/docker | sh
	systemctl start docker
	systemctl enable docker
	docker pull redis
	mkdir -p /usr/local/app/redis /usr/local/app/redis/data
	docker run -p 6379:6379 -v /usr/local/app/redis/data:/data -v /usr/local/app/redis/redis.conf:/etc/redis/redis.conf --name redis -d redis redis-server /etc/redis/redis.conf
	docker pull rabbitmq
	mkdir -p /usr/local/app/rabbitmq
	docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 -v /usr/local/app/rabbitmq/:/var/lib/rabbitmq  rabbitmq:latest

    #curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip
    #unzip protoc-3.13.0-linux-x86_64.zip -d $HOME/.local
    #export PATH="$PATH:$HOME/.local/bin"
    #go get github.com/golang/protobuf/protoc-gen-go
    #go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.0
    #export PATH="$PATH:$(go env GOPATH)/bin"
test:
	go test ./client
	go test ./send_server
	go test ./recv_server