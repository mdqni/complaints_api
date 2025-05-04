FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ./docs ./docs

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app ./
COPY --from=build /app/main ./cmd
COPY --from=build /app/internal ./internal
COPY --from=build /app/docs ./docs

EXPOSE 8080

CMD ["./main"]
