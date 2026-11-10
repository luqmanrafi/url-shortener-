FROM golang:1.27 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /server /app/server

RUN mkdir -p /data

ENV PORT=3000
ENV DB_PATH=/data/urls.db
EXPOSE 3000

CMD [ "/app/server" ]