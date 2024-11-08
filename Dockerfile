FROM golang:1.22-alpine AS build
WORKDIR /app/golang
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/golang/main ./
COPY .env .
EXPOSE 3002
CMD ["./main"]
