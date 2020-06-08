FROM golang:1.14.0-alpine3.11 as rdslogs
RUN apk update && apk add git
WORKDIR /rdslogs
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rdslogs main.go


FROM golang:1.14.0-alpine3.11
WORKDIR /app

COPY --from=rdslogs /rdslogs/rdslogs rdslogs
ENTRYPOINT ["/app/rdslogs"]
