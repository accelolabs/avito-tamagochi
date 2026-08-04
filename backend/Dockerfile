FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api

FROM alpine:3.23

RUN adduser -D -H -u 10001 appuser

COPY --from=build /out/api /usr/local/bin/api

USER appuser
EXPOSE 8080

ENTRYPOINT ["api"]
