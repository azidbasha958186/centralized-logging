FROM golang:alpine
WORKDIR /src
COPY . .
RUN go build -o /app/log-collector

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/log-collector /app/log-collector
EXPOSE 1514/udp 1515 8081
ENV COLLECTOR_UDP_ADDR=:1514
ENV COLLECTOR_TCP_ADDR=:1515
ENV LOG_SERVER_URL=http://log-server:9000/ingest
CMD ["/app/log-collector"]
