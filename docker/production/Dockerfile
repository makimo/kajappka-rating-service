FROM golang:alpine
WORKDIR /go/src/github.com/makimo/kajappka-rating-service/
RUN ls
COPY . .
RUN ls
RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/makimo/kajappka-rating-service/ .
CMD ["./app"]
