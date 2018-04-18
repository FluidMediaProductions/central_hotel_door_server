FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/hotel_gateway
WORKDIR /go/src/hotel_gateway

RUN go get -v -u github.com/golang/dep/...
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./ /go/src/hotel_gateway
RUN CGO_ENABLED=0 go build -o hotel_gateway

FROM scratch

COPY --from=0 /go/src/hotel_gateway/hotel_gateway /hotel_gateway

ENTRYPOINT ["/hotel_gateway"]