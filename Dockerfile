FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o up .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/up .
EXPOSE 8080
CMD ["./up"]
