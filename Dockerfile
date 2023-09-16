FROM golang:1.21 as application
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go mod download
RUN go build -o myapp
RUN chmod +x /app/myapp
CMD ["./myapp"]

FROM alpine:latest
WORKDIR /app
COPY --from=application /app /app/app
CMD ["/app/app"]