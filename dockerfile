FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates
COPY --from=builder /usr/local/bin/app /app

EXPOSE 8080
USER 65534:65534

CMD ["/app"]