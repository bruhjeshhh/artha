FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o user-service ./cmd/user-service && \
    go build -o rental-service ./cmd/rental-service && \
    go build -o grocery-service ./cmd/grocery-service && \
    go build -o transport-service ./cmd/transport-service && \
    go build -o inflation-service ./cmd/inflation-service && \
    go build -o geospatial-service ./cmd/geospatial-service && \
    go build -o cost-prediction-service ./cmd/cost-prediction-service

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/user-service /app/user-service
COPY --from=builder /app/rental-service /app/rental-service
COPY --from=builder /app/grocery-service /app/grocery-service
COPY --from=builder /app/transport-service /app/transport-service
COPY --from=builder /app/inflation-service /app/inflation-service
COPY --from=builder /app/geospatial-service /app/geospatial-service
COPY --from=builder /app/cost-prediction-service /app/cost-prediction-service

# Default: run user-service (overridden by docker-compose per service)
CMD ["./user-service"]
