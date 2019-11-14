# build
FROM golang:1.13.4-alpine as builder
WORKDIR /build
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN apk --no-cache add upx
COPY go.mod go.sum *.go ./
RUN go mod download \
 && go build -a  -ldflags '-s -w' -o expire . \
 && upx --ultra-brute --best --no-progress expire

# final container
FROM scratch
COPY --from=builder /build/expire /
ENTRYPOINT ["/expire"]
