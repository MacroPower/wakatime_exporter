FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o wakatime_exporter .

FROM alpine:latest
COPY --from=builder /app/wakatime_exporter /bin/wakatime_exporter
USER nobody
ENTRYPOINT ["/bin/wakatime_exporter"]
EXPOSE 9212
