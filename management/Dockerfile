FROM golang:1.10-alpine

COPY . /go/src/management
WORKDIR /go/src/management
RUN CGO_ENABLED=0 go build -o management

FROM scratch

COPY --from=0 /go/src/management/management /management

ENTRYPOINT ["/management"]