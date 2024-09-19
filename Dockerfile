FROM golang:latest AS build

WORKDIR /app

COPY . .

RUN go build -o bin/main .

EXPOSE 8081

CMD ["./bin/main"]
