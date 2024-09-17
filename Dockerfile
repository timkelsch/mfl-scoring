FROM golang:1.23.1-alpine AS BUILD
WORKDIR /app
COPY mfl-scoring/go.mod mfl-scoring/go.sum ./
RUN go mod download
COPY mfl-scoring/*.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o main main.go

# Copy artifacts to a clean image
FROM alpine:3.20.3
WORKDIR /app
COPY --from=build /app/main ./main
ENTRYPOINT [ "/app/main" ]