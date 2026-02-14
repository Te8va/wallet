FROM golang:latest AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/wallet/bin/main ./cmd/wallet/

FROM alpine:latest
WORKDIR /wallet
RUN mkdir /wallet/logs
COPY --from=build /build/migrations /wallet/migrations
COPY --from=build /build/cmd/wallet/bin/main .
CMD ["/wallet/main"]