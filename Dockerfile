FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/ ./web/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=build /app/app .
COPY --from=build /app/web ./web

RUN chown -R app:app /app
USER app

EXPOSE 8080
CMD ["./app"]
