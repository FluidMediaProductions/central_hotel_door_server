FROM golang:1.10-alpine

COPY . /go/src/bookings
WORKDIR /go/src/bookings
RUN CGO_ENABLED=0 go build -o bookings

FROM scratch

COPY --from=0 /go/src/bookings/bookings /bookings

ENTRYPOINT ["/bookings"]