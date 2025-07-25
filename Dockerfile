FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./out/server ./cmd/main.go

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/public ./public
COPY --from=builder /app/out/server .

EXPOSE 8080

CMD [ "./server" ]