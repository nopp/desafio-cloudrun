FROM golang:1.23.1 AS builder
COPY . .
RUN unset GOPATH \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM golang:latest
LABEL maintainer="Carlos Augusto Malucelli <camalucelli@gmail.com>"
COPY --from=builder /go/main .
RUN chmod +x main
ENTRYPOINT ["./main"]
