FROM golang:1.23-alpine

RUN apk add --no-cache bash curl

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
WORKDIR /app
COPY . .
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh
RUN go build -o main .

ENTRYPOINT ["/wait-for-it.sh", "postgres:5432", "-t", "15", "--", \
           "sh", "-c", "goose -dir /app/db/migration postgres \"$DB_SOURCE\" up && ./main"]