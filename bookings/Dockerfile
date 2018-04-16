FROM golang:1.10-alpine
RUN apk --no-cache add git

RUN mkdir -p /go/src/bookings
WORKDIR /go/src/bookings

RUN go get -v -u github.com/golang/dep/...
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./ /go/src/bookings
RUN CGO_ENABLED=0 go build -o bookings

FROM scratch

COPY --from=0 /go/src/bookings/bookings /bookings

ENTRYPOINT ["/bookings"]