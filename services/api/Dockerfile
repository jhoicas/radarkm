FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server


FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/server /app/server
COPY migrations /app/migrations
ENV RADARKM_API_ADDR=:8080
EXPOSE 8080
CMD ["/app/server"]
