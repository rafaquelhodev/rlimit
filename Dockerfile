FROM golang:1.22.5

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /rlimit

EXPOSE 4000

ENTRYPOINT ["/rlimit", "4000"]