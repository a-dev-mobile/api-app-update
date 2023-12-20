FROM golang:1.21.4

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd
RUN go build -o main .

EXPOSE 443

CMD ["./main"]
