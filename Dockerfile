FROM golang:1.14-alpine3.12 as rdslogs

WORKDIR /rdslogs

COPY go.mod go.sum ./

RUN apk add --no-cache bash git && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o rdslogs main.go

ENTRYPOINT ["/bin/bash"]
