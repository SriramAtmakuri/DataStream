FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.18
RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
