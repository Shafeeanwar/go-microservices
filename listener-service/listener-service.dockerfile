FROM golang:1.18-alpine as builder
RUN mkdir /app
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o listenerApp .
RUN chmod +x /app/listenerApp


FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/listenerApp .
CMD ["/app/listenerApp"]
