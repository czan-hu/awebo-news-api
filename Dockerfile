FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/awebo-api ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/awebo-api ./awebo-api
COPY app/infrastructure/database/migrations ./app/infrastructure/database/migrations

EXPOSE 8080
ENTRYPOINT ["./awebo-api"]
