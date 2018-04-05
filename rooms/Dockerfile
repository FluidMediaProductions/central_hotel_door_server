FROM golang:1.10-alpine

COPY . /go/src/rooms
WORKDIR /go/src/rooms
RUN CGO_ENABLED=0 go build -o rooms

FROM scratch

COPY --from=0 /go/src/rooms/rooms /rooms

ENTRYPOINT ["/rooms"]