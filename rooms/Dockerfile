FROM golang:1.10-alpine

COPY . /go/src/hotels
WORKDIR /go/src/hotels
RUN CGO_ENABLED=0 go build -o hotels main.go

FROM scratch

COPY --from=0 /go/src/hotels/hotels /hotels

ENTRYPOINT ["/hotels"]