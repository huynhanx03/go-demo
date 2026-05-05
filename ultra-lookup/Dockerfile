FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/ultra-lookup-server ./cmd/server
RUN go build -o /out/ultra-lookup-loadgen ./cmd/loadgen

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out/ultra-lookup-server /app/ultra-lookup-server
COPY --from=build /out/ultra-lookup-loadgen /app/ultra-lookup-loadgen

EXPOSE 8080
ENTRYPOINT ["/app/ultra-lookup-server"]
