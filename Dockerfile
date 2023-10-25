FROM golang:1.21-alpine3.18 AS BUILD
# ARG TARGETARCH
# ARG TARGETOS
WORKDIR /app
# Copy dependencies list
COPY mfl-scoring/go.mod mfl-scoring/go.sum ./
RUN ls -al && which go && go mod download
# Build with optional lambda.norpc tag
COPY mfl-scoring/*.go .
# RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
#   go build -tags lambda.norpc -o main main.go
RUN CGO_ENABLED=0 GOOS=linux go build -tags lambda.norpc -o main main.go

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
WORKDIR /app
COPY --from=build /app/main ./main
ENTRYPOINT [ "/app/main" ]