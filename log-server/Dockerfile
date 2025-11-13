FROM golang:alpine
WORKDIR /src
COPY . .
RUN go build -o /app/log-server

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/log-server /app/log-server
RUN mkdir -p /app/data
EXPOSE 9000
ENV SERVER_ADDR=:9000
ENV STORAGE_FILE=/app/data/logs.jsonl
CMD ["/app/log-server"]
