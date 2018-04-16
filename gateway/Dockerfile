FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/gateway
WORKDIR /go/src/gateway

RUN go get -v -u github.com/golang/dep/...
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./ /go/src/gateway
RUN CGO_ENABLED=0 go build -o gateway

FROM scratch

COPY --from=0 /go/src/gateway/gateway /gateway

ENTRYPOINT ["/gateway"]