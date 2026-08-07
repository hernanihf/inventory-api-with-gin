FROM golang:1.26-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init
RUN CGO_ENABLED=0 GOOS=linux go build -o inventory-api .

FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/inventory-api .

EXPOSE 8080
CMD ["./inventory-api"]