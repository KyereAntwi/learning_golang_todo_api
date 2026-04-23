#stage 1: build the go binary

FROM golang:1.26-alpine AS builder
WORKDIR /app

# leverage docker cache by downloading dependencies first
COPY go.mod go.sum ./
RUN go mod download

# copy the full source tree and build the binaries
COPY src ./src
COPY docs ./docs
WORKDIR /app/src/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api .
WORKDIR /app/src/workers
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/workers .

#stage 2: create a minimal runtime image
FROM alpine:latest
WORKDIR /root/

# copy only the compiled binaries from the builder stage
COPY --from=builder /app/bin/api ./api
COPY --from=builder /app/bin/workers ./workers

# expose the API port
EXPOSE 8080

# run both the API and workers when the container starts
CMD ["sh", "-c", "./workers & ./api"]