FROM golang:1.10-alpine

COPY . /go/src/gateway
WORKDIR /go/src/gateway
RUN CGO_ENABLED=0 go build -o gateway

FROM scratch

COPY --from=0 /go/src/gateway/gateway /gateway

ENTRYPOINT ["/gateway"]