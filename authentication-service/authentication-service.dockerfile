FROM golang:1.18-alpine as builder
RUN mkdir /app
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o authApp ./cmd/api
RUN chmod +x /app/authApp


FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/authApp .
CMD ["/app/authApp"]
