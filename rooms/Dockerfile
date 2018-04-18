FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/rooms
WORKDIR /go/src/rooms

RUN go get -v -u github.com/golang/dep/...
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./ /go/src/rooms
RUN CGO_ENABLED=0 go build -o rooms

FROM scratch

COPY --from=0 /go/src/rooms/rooms /rooms

ENTRYPOINT ["/rooms"]