FROM golang:1.18-alpine as builder
RUN mkdir /app
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o mailerServiceApp ./cmd/api
RUN chmod +x /app/mailerServiceApp


FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/mailerServiceApp .
COPY ./templates ./../templates 
COPY ./templates ./templates 
CMD ["/app/mailerServiceApp"]
