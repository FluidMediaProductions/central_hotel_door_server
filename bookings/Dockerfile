FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/bookings
WORKDIR /go/src/bookings

RUN go get -v -d github.com/dgraph-io/dgo
RUN go get -v -d github.com/gorilla/mux
RUN go get -v -d github.com/spf13/viper
RUN go get -v -d google.golang.org/grpc

COPY ./ /go/src/bookings
RUN go get -v -d
RUN CGO_ENABLED=0 go build -o bookings

FROM scratch

COPY --from=0 /go/src/bookings/bookings /bookings

ENTRYPOINT ["/bookings"]