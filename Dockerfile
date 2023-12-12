FROM golang:1.21-alpine AS BUILD
WORKDIR /app
COPY mfl-scoring/go.mod mfl-scoring/go.sum ./
RUN go mod download
COPY mfl-scoring/*.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o main main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main ./main
ENTRYPOINT [ "/app/main" ]