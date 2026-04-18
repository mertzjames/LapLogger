FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o laplogger .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget \
	&& addgroup -S app \
	&& adduser -S -G app app
WORKDIR /app
COPY --from=builder /app/laplogger .
RUN chown app:app /app/laplogger
USER app
EXPOSE 8080
CMD ["./laplogger"]
