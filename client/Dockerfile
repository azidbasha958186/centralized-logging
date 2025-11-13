FROM golang:alpine
WORKDIR /src
COPY . .
RUN go build -o /app/client

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/client /app/client
CMD ["/app/client"]
