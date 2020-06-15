FROM golang:1.14-alpine3.12 as rdslogs

ENV GOPATH=/go

RUN mkdir -p /go/src/github.com/razorpay/rdslogs && \
    apk add --no-cache git

WORKDIR /go/src/github.com/razorpay/rdslogs/

COPY go.mod go.sum /go/src/github.com/razorpay/rdslogs/

RUN go mod download

COPY . /go/src/github.com/razorpay/rdslogs/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o rdslogs main.go


FROM golang:1.14-alpine3.12

WORKDIR /app

COPY --from=rdslogs /go/src/github.com/razorpay/rdslogs/rdslogs rdslogs

ENTRYPOINT ["/app/rdslogs"]
