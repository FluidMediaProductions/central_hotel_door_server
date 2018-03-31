FROM golang:1.10-alpine

COPY . /go/src/auth
WORKDIR /go/src/auth
RUN CGO_ENABLED=0 go build -o auth main.go

FROM scratch

COPY --from=0 /go/src/auth/auth /auth

ENTRYPOINT ["/auth"]