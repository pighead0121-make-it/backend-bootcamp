FROM golang:1.26.3 AS builder

WORKDIR /app

COPY . .

RUN go build -o book-api .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/book-api .

CMD [ "./book-api" ]