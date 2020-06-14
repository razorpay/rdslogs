FROM golang:1.14-alpine3.12 as rdslogs

ENV GOPATH=/go

RUN mkdir -p /go/src/github.com/razorpay/rdslogs && apk add --no-cache bash git mariadb-client

WORKDIR /go/src/github.com/razorpay/rdslogs/

COPY . /go/src/github.com/razorpay/rdslogs/

RUN rm -f go.mod go.sum && go mod init && go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o rdslogs main.go

ENTRYPOINT ["sleep", "864000"]
