FROM golang:1.16
RUN mkdir /app
ADD . /app
WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/cmd/server/server /usr/bin/
CMD ["sh", "-c", "/usr/bin/server", "-queue", "$QUEUE", "-out", "$OUT"]
