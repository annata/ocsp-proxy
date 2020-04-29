FROM golang:alpine

ENV GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY main.go .
RUN go build -o main main.go
WORKDIR /dist
RUN cp /build/main .
CMD ["/dist/main"]
